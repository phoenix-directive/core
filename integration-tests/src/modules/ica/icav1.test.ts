import { AccAddress } from "@terra-money/feather.js";
import { DirectSecp256k1HdWallet, EncodeObject, Registry } from "@cosmjs/proto-signing";
import { defaultRegistryTypes, DeliverTxResponse, SigningStargateClient } from "@cosmjs/stargate";
import { stringToPath } from "@cosmjs/crypto";
import { blockInclusion, getEventsByIndex, getLCDClient, getMnemonics } from "../../helpers";
import { MsgSend } from "cosmjs-types/cosmos/bank/v1beta1/tx";
import { MsgTransfer } from "cosmjs-types/ibc/applications/transfer/v1/tx";
import { MsgRegisterInterchainAccount, MsgSendTx } from "cosmjs-types/ibc/applications/interchain_accounts/controller/v1/tx";
import { CosmosTx, InterchainAccountPacketData, Type } from "cosmjs-types/ibc/applications/interchain_accounts/v1/packet";
import { Order } from "cosmjs-types/ibc/core/channel/v1/channel";

const CHAIN_1_RPC = "http://localhost:16657";
const CHAIN_1_ID = "test-1";
const TERRA_HD_PATH = "m/44'/330'/0'/0/0";
const COSMJS_FEE = {
    amount: [{ denom: "uluna", amount: "22500" }],
    gas: "900000", };

describe("ICA Module", () => {
    // Prepare environment clients, accounts and wallets
    const LCD = getLCDClient();
    const { icaMnemonic } = getMnemonics();
    const externalAccAddr = icaMnemonic.accAddress("terra");
    let ibcCoinDenom: string | undefined;
    let intechainAccountAddr: AccAddress | undefined;
    let chain1Cosmjs: SigningStargateClient;
    let cosmjsRegistry: Registry;

    beforeAll(async () => {
        cosmjsRegistry = new Registry(defaultRegistryTypes);
        cosmjsRegistry.register(MsgRegisterInterchainAccount.typeUrl, MsgRegisterInterchainAccount as any);
        cosmjsRegistry.register(MsgTransfer.typeUrl, MsgTransfer as any);
        cosmjsRegistry.register(MsgSendTx.typeUrl, MsgSendTx as any);
        cosmjsRegistry.register(MsgSend.typeUrl, MsgSend as any);

        const signer = await DirectSecp256k1HdWallet.fromMnemonic(icaMnemonic.mnemonic, {
            prefix: "terra",
            hdPaths: [stringToPath(TERRA_HD_PATH)],
        });
        const [account] = await signer.getAccounts();
        expect(account.address).toStrictEqual(externalAccAddr);

        chain1Cosmjs = await SigningStargateClient.connectWithSigner(CHAIN_1_RPC, signer, {
            registry: cosmjsRegistry,
        });
    });

    async function signAndBroadcastCosmjs(messages: EncodeObject[]): Promise<DeliverTxResponse> {
        return chain1Cosmjs.signAndBroadcast(externalAccAddr, messages, COSMJS_FEE);
    }

    const waitForInterchainAccount = async () => {
        for (let i = 0; i <= 10; i++) {
            await blockInclusion();
            const res = await LCD.chain1.icaV1.controllerAccountAddress(externalAccAddr, "connection-0")
                .catch((e) => {
                    const expectMsg = "failed to retrieve account address for icacontroller-";
                    expect(e.response.data.message.startsWith(expectMsg)).toBeTruthy();
                });
            if (res) {
                expect(res.address).toBeDefined();
                intechainAccountAddr = res.address;
                return;
            }
        }
    }

    test('Must contain the expected module params', async () => {
        // Query ica host module params
        const hostResParams = await LCD.chain2.icaV1.hostParams("test-2");
        expect(hostResParams.params)
            .toStrictEqual({
                "host_enabled": true,
                "allow_messages": [
                    "/cosmos.authz.v1beta1.MsgExec",
                    "/cosmos.authz.v1beta1.MsgGrant",
                    "/cosmos.authz.v1beta1.MsgRevoke",
                    "/cosmos.bank.v1beta1.MsgSend",
                    "/cosmos.bank.v1beta1.MsgMultiSend",
                    "/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
                    "/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommission",
                    "/cosmos.distribution.v1beta1.MsgFundCommunityPool",
                    "/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
                    "/cosmos.feegrant.v1beta1.MsgGrantAllowance",
                    "/cosmos.feegrant.v1beta1.MsgRevokeAllowance",
                    "/cosmos.gov.v1beta1.MsgVoteWeighted",
                    "/cosmos.gov.v1beta1.MsgSubmitProposal",
                    "/cosmos.gov.v1beta1.MsgDeposit",
                    "/cosmos.gov.v1beta1.MsgVote",
                    "/cosmos.staking.v1beta1.MsgEditValidator",
                    "/cosmos.staking.v1beta1.MsgDelegate",
                    "/cosmos.staking.v1beta1.MsgUndelegate",
                    "/cosmos.staking.v1beta1.MsgBeginRedelegate",
                    "/cosmos.staking.v1beta1.MsgCreateValidator",
                    "/cosmos.vesting.v1beta1.MsgCreateVestingAccount",
                    "/ibc.applications.transfer.v1.MsgTransfer",
                    "/cosmwasm.wasm.v1.MsgStoreCode",
                    "/cosmwasm.wasm.v1.MsgInstantiateContract",
                    "/cosmwasm.wasm.v1.MsgExecuteContract",
                    "/cosmwasm.wasm.v1.MsgMigrateContract"
                ]
            });

        // Query contoller module params
        const controllerResParams = await LCD.chain2.icaV1.controllerParams("test-2");
        expect(controllerResParams.params)
            .toStrictEqual({
                controller_enabled: true,
            });
    });

    test('Must query the interchain account to determine its existance', async () => {
        // Query the account address of the interchain account
        let res = await LCD.chain1.icaV1.controllerAccountAddress(externalAccAddr, "connection-0")
            .catch(e => {
                // assert that the expected error is that it failed to retreive the account
                const expectMsg = "failed to retrieve account address for icacontroller-";
                expect(e.response.data.message.startsWith(expectMsg)).toBeTruthy();
            })

        // if res is defined then the account exists
        if (res !== undefined) {
            expect(res.address).toBeDefined();
            intechainAccountAddr = res.address;
            // Check during 5 blocks for the receival 
            // of the IBC coin on chain-2
            for (let i = 0; i <= 5; i++) {
                await blockInclusion();
                let _ibcCoin = (await LCD.chain2.bank.balance(intechainAccountAddr))[0].find(c => c.denom.startsWith("ibc/"));
                if (_ibcCoin) {
                    expect(_ibcCoin.denom.startsWith("ibc/")).toBeTruthy();
                    ibcCoinDenom = _ibcCoin.denom
                    break;
                }
            }
        }
    });

    test('Must create the interchain account if des not already exist', async () => {
        const registerMsg: EncodeObject = {
            typeUrl: MsgRegisterInterchainAccount.typeUrl,
            value: MsgRegisterInterchainAccount.fromPartial({
                owner: externalAccAddr,
                connectionId: "connection-0",
                version: "",
                ordering: Order.ORDER_ORDERED,
            }),
        };
        const txResult = await signAndBroadcastCosmjs([registerMsg]);

        if (txResult.code !== 0) {
            expect(txResult.rawLog ?? "").toContain("existing active channel");
            expect(txResult.rawLog ?? "").toContain("active channel already set for this owner");
        } else {
            expect(txResult.code).toStrictEqual(0);
            const lcdTxResult = await LCD.chain1.tx.txInfo(txResult.transactionHash, CHAIN_1_ID) as any;
            const events = getEventsByIndex(lcdTxResult.events, 0);

            expect(events[0].type).toStrictEqual("message");
            expect(events[0].attributes).toEqual(expect.arrayContaining([{
                "index": true,
                "key": "action",
                "value": "/ibc.applications.interchain_accounts.controller.v1.MsgRegisterInterchainAccount"
            }, {
                "index": true,
                "key": "sender",
                "value": "terra1p4kcrttuxj9kyyvv5px5ccgwf0yrw74yp7jqm6"
            }, {
                "index": true,
                "key": "msg_index",
                "value": "0"
            }]));
            expect(events[2].type).toStrictEqual("message");
            expect(events[2].attributes).toEqual(expect.arrayContaining([{
                "index": true,
                "key": "module",
                "value": "ibc_channel"
            }, {
                "index": true,
                "key": "msg_index",
                "value": "0"
            }]));

            // Check during 15 blocks for the ICA handshake to finish.
            for (let i = 0; i <= 15; i++) {
                await blockInclusion();
                let res = await LCD.chain1.icaV1.controllerAccountAddress(externalAccAddr, "connection-0")
                    .catch((e) => {
                        const expectMsg = "failed to retrieve account address for icacontroller-";
                        expect(e.response.data.message.startsWith(expectMsg)).toBeTruthy();
                    })
                if (res) {
                    expect(res.address).toBeDefined();
                    intechainAccountAddr = res.address;
                    break;
                }
            }
        }

        await waitForInterchainAccount();
        expect(intechainAccountAddr).toBeDefined();
    });

    describe('After assuring the interchain account exists', () => {
        test("Must send funds to the interchain account from chain-1 to chain-2", async () => {
            if (typeof intechainAccountAddr === "string") {
                let blockHeight = (await LCD.chain1.tendermint.blockInfo("test-1")).block.header.height;
                const transferMsg: EncodeObject = {
                    typeUrl: MsgTransfer.typeUrl,
                    value: MsgTransfer.fromPartial({
                        sourcePort: "transfer",
                        sourceChannel: "channel-0",
                        token: { denom: "uluna", amount: "100000000" },
                        sender: externalAccAddr,
                        receiver: intechainAccountAddr,
                        timeoutHeight: {
                            revisionNumber: 2n,
                            revisionHeight: BigInt(parseInt(blockHeight) + 100),
                        },
                        timeoutTimestamp: 0n,
                        memo: "",
                        encoding: "",
                    }),
                };

                const txResult = await signAndBroadcastCosmjs([transferMsg]);
                expect(txResult.code).toStrictEqual(0);
                expect(txResult).toBeDefined();
                // Check during 5 blocks for the receival 
                // of the IBC coin on chain-2
                for (let i = 0; i <= 5; i++) {
                    await blockInclusion();
                    let _ibcCoin = (await LCD.chain2.bank.balance(intechainAccountAddr))[0].find(c => c.denom.startsWith("ibc/"));
                    if (_ibcCoin) {
                        expect(_ibcCoin.denom.startsWith("ibc/")).toBeTruthy();
                        ibcCoinDenom = _ibcCoin.denom
                        break;
                    }
                }
            } else {
                // This case should never happen but if something goes wrong
                // this is a check to fail.
                expect(intechainAccountAddr).toBeDefined()
            }
        });

        test("Must control the interchain account from chain-1 to send funds on chain-2 from the account address to a burnAddress", async () => {
            const burnAddress = "terra1zdpgj8am5nqqvht927k3etljyl6a52kwqup0je";
            const hostChainMsgSend: EncodeObject = {
                typeUrl: MsgSend.typeUrl,
                value: MsgSend.fromPartial({
                    fromAddress: intechainAccountAddr as string,
                    toAddress: burnAddress,
                    amount: [{ denom: ibcCoinDenom, amount: "1000" }],
                }),
            };
            const interchainAccountPacketData = InterchainAccountPacketData.fromPartial({
                type: Type.TYPE_EXECUTE_TX,
                data: CosmosTx.encode(CosmosTx.fromPartial({
                    messages: [cosmjsRegistry.encodeAsAny(hostChainMsgSend)],
                })).finish(),
                memo: "",
            });
            const msgSendTx: EncodeObject = {
                typeUrl: MsgSendTx.typeUrl,
                value: MsgSendTx.fromPartial({
                    owner: externalAccAddr,
                    connectionId: "connection-0",
                    relativeTimeout: BigInt(Date.now()) * 1000000n + 600000000n,
                    packetData: interchainAccountPacketData,
                }),
            };

            const broadcastResult = await signAndBroadcastCosmjs([msgSendTx]);
            expect(broadcastResult.code).toStrictEqual(0);
            let txResult = await LCD.chain1.tx.txInfo(broadcastResult.transactionHash, CHAIN_1_ID) as any;
            const events = getEventsByIndex(txResult.events, 0);
            expect(events[0])
                .toStrictEqual({
                    "type": "message",
                    "attributes": [{
                        "index": true,
                        "key": "action",
                        "value": "/ibc.applications.interchain_accounts.controller.v1.MsgSendTx"
                    }, {
                        "index": true,
                        "key": "sender",
                        "value": "terra1p4kcrttuxj9kyyvv5px5ccgwf0yrw74yp7jqm6"
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
            expect(events[2])
                .toStrictEqual({
                    "type": "message",
                    "attributes": [{
                        "index": true,
                        "key": "module",
                        "value": "ibc_channel"
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                })


            // Check during 5 blocks for the receival 
            // of the IBC coin on chain-2
            for (let i = 0; i <= 5; i++) {
                await blockInclusion();
                const bankRes = await LCD.chain2.bank.balance(burnAddress);
                const coins = bankRes[0].find(c => c.denom === ibcCoinDenom);
                if (coins) {
                    expect(coins).toBeDefined();
                    expect(coins?.denom).toStrictEqual(ibcCoinDenom);
                    expect(coins?.amount.toNumber()).toBeGreaterThanOrEqual(1000);
                    break;
                }
            }
        })
    });
});
