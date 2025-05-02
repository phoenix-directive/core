
import { MnemonicKey, MsgExecuteContract, MsgInstantiateContract, MsgStoreCode } from "@terra-money/feather.js";
import { getMnemonics, getLCDClient, blockInclusion, ibcTransfer, getEventsByIndex } from "../../helpers";
import fs from "fs";
import path from 'path';
import { execSync, exec } from 'child_process';

describe("Wasm Module (https://github.com/CosmWasm/wasmd/releases/tag/v0.45.0) ", () => {
    // Prepare environment clients, accounts and wallets
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    const wallet = LCD.chain1.wallet(accounts.wasmContracts);
    const walletAddress = accounts.wasmContracts.accAddress("terra");
    const randomWalletAddress = new MnemonicKey().accAddress("terra");
    let cw20BaseCodeId: number;
    let ics20CodeId: number;
    let cw20ContractAddr: string;
    let ics20ContractAddr: string;
    let ics20ContractChannelId: string;

    // Validate that wasm module has the correct params
    test('Must have the correct module params', async () => {
        const moduleParams = await LCD.chain1.wasm.params("test-1");

        expect(moduleParams)
            .toStrictEqual({
                params: {
                    "code_upload_access": {
                        "addresses": [],
                        "permission": "Everybody"
                    },
                    "instantiate_default_permission": "Everybody",
                }
            });
    })

    // Validate that wasm module has the correct params
    test('Must deploy *cw20_base* and *cw20_ics20* contracts', async () => {
        let tx = await wallet.createAndSignTx({
            msgs: [
                new MsgStoreCode(walletAddress, fs.readFileSync(path.join(__dirname, "/../../contracts/cw20_base.wasm")).toString("base64")),
                new MsgStoreCode(walletAddress, fs.readFileSync(path.join(__dirname, "/../../contracts/cw20_ics20.wasm")).toString("base64"))
            ],
            chainID: "test-1",
        });

        let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();

        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        const events = getEventsByIndex(txResult.events, 0);
        cw20BaseCodeId = Number(events[1].attributes[1].value);
        expect(cw20BaseCodeId).toBeDefined();
        const events2 = getEventsByIndex(txResult.events, 1);
        ics20CodeId = Number(events2[1].attributes[1].value);
        expect(ics20CodeId).toBeDefined();
    })

    describe("after contracts has been deployed", () => {
        test("Must instantiate *cw20_base* and *cw20_ics20* contract", async () => {
            let tx = await wallet.createAndSignTx({
                msgs: [new MsgInstantiateContract(
                    walletAddress,
                    walletAddress,
                    cw20BaseCodeId,
                    {
                        name: "Bitcoin",
                        symbol: "BTC",
                        decimals: 8,
                        initial_balances: [{
                            address: walletAddress,
                            amount: "100000000",
                        }]
                    },
                    undefined,
                    "A cw20 contract" + Math.random(),
                )],
                chainID: "test-1",
            });
            let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            let events = getEventsByIndex(txResult.events, 0);
            cw20ContractAddr = events[1].attributes[0].value;
            expect(cw20ContractAddr).toBeDefined();

            tx = await wallet.createAndSignTx({
                msgs: [new MsgInstantiateContract(
                    walletAddress,
                    walletAddress,
                    ics20CodeId,
                    {
                        default_timeout: 60,
                        gov_contract: cw20ContractAddr,
                        allowlist: [{
                            contract: cw20ContractAddr,
                            gas_limit: 1000000,
                        }],
                        default_gas_limit: 1000000
                    },
                    undefined,
                    "A cw20 contract" + Math.random(),
                )],
                chainID: "test-1",
            });
            result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();
            txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            events = getEventsByIndex(txResult.events, 0);
            ics20ContractAddr = events[1].attributes[0].value;
            expect(ics20ContractAddr).toBeDefined();
        })

        test("must create the channel for the ICS20 contract", async () => {
            // Stop the relayer to don't create conflicts
            try {
                execSync("pkill relayer")
            }
            catch (e) {
                console.log(e)
            }

            // Create the path
            const pathToRelayDir = path.join(__dirname, "/../../test-data/relayer");
            execSync(`relayer tx link "test1-test2" --src-port="wasm.${ics20ContractAddr}" --dst-port="transfer" --version="ics20-1" --home="${pathToRelayDir}"`, { stdio: "ignore" })
            await blockInclusion();

            // Start the relayer again
            const relayerStart = exec(`relayer start "test1-test2" -p="events" -b=100 --flush-interval="1s" --time-threshold="1s" --home="${pathToRelayDir}" > ${pathToRelayDir}/relayer.log 2>&1`)
            relayerStart.unref();

            const res = await LCD.chain1.ibc.channels("test-1", {
                "pagination.limit": 1,
                "pagination.reverse": "true",
            });

            expect(res.channels[0]).toBeDefined();
            expect(res.channels[0].channel_id).toBeDefined();
            ics20ContractChannelId = res.channels[0].channel_id;
        }, 120_000)



        describe("after channel has been created", () => {
            test("Must send funds from test-1 to test-2", async () => {
                // SubMessage to Transfer the funds thoguht the IBC channel
                // which must be parsed to base64 and embeded into the "send"
                // message. (we're not using JSON.stringify(object) because it causes and error)
                let subMsg = Buffer.from(`{
                    "channel": "${ics20ContractChannelId}",
                    "remote_address":"${randomWalletAddress}"
                }`).toString("base64");

                let tx = await wallet.createAndSignTx({
                    msgs: [new MsgExecuteContract(
                        walletAddress,
                        cw20ContractAddr,
                        {
                            send: {
                                contract: ics20ContractAddr,
                                amount: "100000",
                                msg: subMsg
                            }
                        },
                    )],
                    chainID: "test-1",
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                let events = getEventsByIndex(txResult.events, 0);

                // Asser the order of events execution on chain
                expect(events[0]).toStrictEqual({
                    "type": "message",
                    "attributes": [{
                        "index": true,
                        "key": "action",
                        "value": "/cosmwasm.wasm.v1.MsgExecuteContract"
                    }, {
                        "index": true,
                        "key": "sender",
                        "value": walletAddress
                    }, {
                        "index": true,
                        "key": "module",
                        "value": "wasm"
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
                expect(events[1]).toStrictEqual({
                    "type": "execute",
                    "attributes": [{
                        "index": true,
                        "key": "_contract_address",
                        "value": cw20ContractAddr
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
                expect(events[2]).toStrictEqual({
                    "type": "wasm",
                    "attributes": [{
                        "index": true,
                        "key": "_contract_address",
                        "value": cw20ContractAddr
                    }, {
                        "index": true,
                        "key": "action",
                        "value": "send"
                    }, {
                        "index": true,
                        "key": "from",
                        "value": walletAddress
                    }, {
                        "index": true,
                        "key": "to",
                        "value": ics20ContractAddr
                    }, {
                        "index": true,
                        "key": "amount",
                        "value": "100000"
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
                expect(events[3]).toStrictEqual({
                    "type": "execute",
                    "attributes": [{
                        "index": true,
                        "key": "_contract_address",
                        "value": ics20ContractAddr
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
                expect(events[4]).toStrictEqual({
                    "type": "wasm",
                    "attributes": [{
                        "index": true,
                        "key": "_contract_address",
                        "value": ics20ContractAddr
                    }, {
                        "index": true,
                        "key": "action",
                        "value": "transfer"
                    }, {
                        "index": true,
                        "key": "amount",
                        "value": "100000"
                    }, {
                        "index": true,
                        "key": "denom",
                        "value": "cw20:" + cw20ContractAddr
                    }, {
                        "index": true,
                        "key": "receiver",
                        "value": randomWalletAddress
                    }, {
                        "index": true,
                        "key": "sender",
                        "value": walletAddress
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });

                // Assert the assets reach the other chain with the correct amount
                await ibcTransfer();
                await ibcTransfer();
                const queryRes = await LCD.chain2.bank.balance(randomWalletAddress);
                let ibcCoin = queryRes[0].find(coin => coin.denom.startsWith("ibc/"));
                expect(ibcCoin).toBeDefined();
                expect(ibcCoin?.amount?.toString()).toStrictEqual("100000");
            })
        })
    })
});