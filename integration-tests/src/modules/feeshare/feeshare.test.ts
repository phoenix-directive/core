import { getMnemonics, blockInclusion, getLCDClient, getValueByIndexAndTypeAndKey, getEventsByIndex, signAndBroadcastTx } from "../../helpers";
import { Coins, MnemonicKey, MsgExecuteContract, MsgInstantiateContract, MsgRegisterFeeShare, MsgStoreCode } from "@terra-money/feather.js";
import fs from "fs";
import path from 'path';

describe("Feeshare Module (https://github.com/terra-money/core/tree/release/v2.6/x/feeshare) ", () => {
    // Prepare environment clients, accounts and wallets
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    const wallet = LCD.chain1.wallet(accounts.feeshareMnemonic);
    const feeshareAccountAddress = accounts.feeshareMnemonic.accAddress("terra");
    const randomAccountAddress = new MnemonicKey().accAddress("terra");
    let contractAddress: string;

    // Read the reflect contract, store on chain, 
    // instantiate to be used in the following tests
    // and finally save the contract address.
    beforeAll(async () => {
        let result = await signAndBroadcastTx(wallet, {
            msgs: [new MsgStoreCode(
                feeshareAccountAddress,
                fs.readFileSync(path.join(__dirname, "/../../contracts/reflect.wasm")).toString("base64"),
            )],
            chainID: "test-1",
        });

        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        let codeId = Number(getValueByIndexAndTypeAndKey(txResult.events, 0, "store_code", "code_id"));
        expect(codeId).toBeDefined();

        const msgInstantiateContract = new MsgInstantiateContract(
            feeshareAccountAddress,
            feeshareAccountAddress,
            codeId,
            {},
            Coins.fromString("1uluna"),
            "Reflect contract " + Math.random(),
        );

        result = await signAndBroadcastTx(wallet, {
            msgs: [msgInstantiateContract],
            chainID: "test-1",
        });
        await blockInclusion();
        txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        contractAddress = getValueByIndexAndTypeAndKey(txResult.events, 0, "instantiate", "_contract_address");
        expect(contractAddress).toBeDefined();
    });

    test('Must contain the expected module params', async () => {
        // Query feeshare module params
        const moduleParams = await LCD.chain1.feeshare.params("test-1");

        expect(moduleParams)
            .toMatchObject({
                "params": {
                    "enable_fee_share": true,
                    "developer_shares": "0.500000000000000000",
                    "allowed_denoms": []
                }
            });
    });

    test('Must register fee share', async () => {
        // Register feeshare
        let result = await signAndBroadcastTx(wallet, {
            msgs: [new MsgRegisterFeeShare(
                contractAddress,
                feeshareAccountAddress,
                randomAccountAddress,
            )],
            chainID: "test-1",
        });

        await blockInclusion();

        // Check the tx logs
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        let txEvents = getEventsByIndex(txResult.events, 0);
        expect(txEvents)
            .toMatchObject([{
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "action",
                    "value": "/juno.feeshare.v1.MsgRegisterFeeShare"
                }, {
                    "index": true,
                    "key": "sender",
                    "value": feeshareAccountAddress,
                }, {
                    "index": true,
                    "key": "module",
                    "value": "feeshare"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "register_feeshare",
                "attributes": [{
                    "index": true,
                    "key": "contract",
                    "value": contractAddress
                }, {
                    "index": true,
                    "key": "withdrawer_address",
                    "value": randomAccountAddress,
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            }])

        // Check the registered feeshares by contractAddress
        let feesharesBy = await LCD.chain1.feeshare.feeshares("test-1", contractAddress);
        expect(feesharesBy)
            .toMatchObject({
                "feeshare": {
                    "contract_address": contractAddress,
                    "deployer_address": feeshareAccountAddress,
                    "withdrawer_address": randomAccountAddress,
                }
            })
        // Check that querying all feeshares returns at least one feeshares
        let feesharesByWallet = await LCD.chain1.feeshare.feeshares("test-1");
        expect(feesharesByWallet.feeshare.length).toBeGreaterThan(0);
        await blockInclusion();

        // Send an execute message to the reflect contract
        let msgExecute = new MsgExecuteContract(
            feeshareAccountAddress,
            contractAddress,
            {
                change_owner: {
                    owner: randomAccountAddress,
                }
            },
        );
        result = await signAndBroadcastTx(wallet, {
            msgs: [msgExecute],
            chainID: "test-1",
            fee: {
                amount: [{ denom: "uluna", amount: "400000" }],
                gas: "200000",
            },
        });
        await blockInclusion();

        // Check the tx logs have the expected events
        txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        txEvents = getEventsByIndex(txResult.events, 0);
        expect(txEvents)
            .toMatchObject([{
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "action",
                    "value": "/cosmwasm.wasm.v1.MsgExecuteContract"
                }, {
                    "index": true,
                    "key": "sender",
                    "value": feeshareAccountAddress
                }, {
                    "index": true,
                    "key": "module",
                    "value": "wasm"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "execute",
                "attributes": [{
                    "index": true,
                    "key": "_contract_address",
                    "value": contractAddress
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "wasm",
                "attributes": [{
                    "index": true,
                    "key": "_contract_address",
                    "value": contractAddress
                }, {
                    "index": true,
                    "key": "action",
                    "value": "change_owner"
                }, {
                    "index": true,
                    "key": "owner",
                    "value": randomAccountAddress
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            }
            ])
        await blockInclusion()

        // Query the random account (new owner of the contract)
        // and validate that the account has received 50% of the fees
        const bankAmount = await LCD.chain1.bank.balance(randomAccountAddress);
        expect(bankAmount[0])
            .toMatchObject(Coins.fromString("200000uluna"))
    });
});
