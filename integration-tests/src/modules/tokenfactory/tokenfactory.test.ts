import { Coin, Coins, Fee, MnemonicKey, MsgBurn, MsgChangeAdmin, MsgCreateDenom, MsgInstantiateContract, MsgMint, MsgStoreCode, MsgSetBeforeSendHook, MsgSend, MsgSubmitProposal, MsgUpdateParamsTokenFactory, MsgVote } from "@terra-money/feather.js";
import { getMnemonics, getLCDClient, blockInclusion, votingPeriod } from "../../helpers";
import fs from "fs";
import path from 'path';
import { VoteOption } from "../../../../../terra.proto/js/cosmos/gov/v1/gov";

describe("TokenFactory Module (https://github.com/terra-money/core/tree/release/v2.7/x/tokenfactory) ", () => {
    // Prepare environment clients, accounts and wallets
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    const wallet = LCD.chain1.wallet(accounts.tokenFactoryMnemonic);
    const tokenFactoryWalletAddr = accounts.tokenFactoryMnemonic.accAddress("terra");
    const randomAccountAddr = new MnemonicKey().accAddress("terra");
    let contractAddress: string;
    let subdenom = Math.random().toString(36).substring(7);
    let factoryDenom: string | undefined = undefined
    let customQueryContractAddress: string;
    let codeId: number;
    const govAddress = "terra10d07y265gmmuvt4z0w9aw880jnsr700juxf95n"
    const val1Wallet = LCD.chain1.wallet(accounts.val1);
    const val1WalletAddress = val1Wallet.key.accAddress("terra");

    // Read the no100 contract, store on chain, 
    // instantiate to be used in the following tests
    // and finally save the contract address.
    beforeAll(async () => {
        let tx = await wallet.createAndSignTx({
            msgs: [new MsgStoreCode(
                tokenFactoryWalletAddr,
                fs.readFileSync(path.join(__dirname, "/../../contracts/no100.wasm")).toString("base64"),
            )],
            chainID: "test-1",
        });

        let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        codeId = Number(txResult.logs[0].events[1].attributes[1].value);
        expect(codeId).toBeDefined();

        const msgInstantiateContract = new MsgInstantiateContract(
            tokenFactoryWalletAddr,
            tokenFactoryWalletAddr,
            codeId,
            {},
            Coins.fromString("1uluna"),
            "no100 contract " + Math.random(),
        );

        tx = await wallet.createAndSignTx({
            msgs: [msgInstantiateContract],
            chainID: "test-1",
        });
        result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        contractAddress = txResult.logs[0].events[4].attributes[0].value;
        expect(contractAddress).toBeDefined();
    })

    // Validate the token factory having the correct params
    test.skip('Must have the correct module params', async () => {
        const moduleParams = await LCD.chain1.tokenfactory.params("test-1");

        expect(moduleParams)
            .toStrictEqual({
                "params": {
                    "denom_creation_fee": [{
                        "amount": "10000000",
                        "denom": "uluna"
                    }],
                    "denom_creation_gas_consume": "1000000",
                    "whitelisted_hooks": [],
                }
            });
    })

    // Create a denom using token factory,
    // store the factoryDenom and read the 
    // transaction result logs to assert 
    // the logs are correctly formatted.
    test('Must create a denom', async () => {
        let tx = await wallet.createAndSignTx({
            msgs: [
                new MsgCreateDenom(
                    tokenFactoryWalletAddr,
                    subdenom,
                ),
            ],
            chainID: "test-1",
        });
        let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        factoryDenom = txResult.logs[0].eventsByType.create_denom.new_token_denom[0] as string
        expect(txResult.logs[0].events).toStrictEqual([{
            "type": "message",
            "attributes": [{
                "key": "action",
                "value": "/osmosis.tokenfactory.v1beta1.MsgCreateDenom"
            }, {
                "key": "sender",
                "value": tokenFactoryWalletAddr
            }, {
                "key": "module",
                "value": "tokenfactory"
            }]
        }, {
            "type": "coin_spent",
            "attributes": [{
                "key": "spender",
                "value": tokenFactoryWalletAddr
            }, {
                "key": "amount",
                "value": "10000000uluna"
            }]
        }, {
            "type": "coin_received",
            "attributes": [{
                "key": "receiver",
                "value": "terra1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8pm7utl"
            }, {
                "key": "amount",
                "value": "10000000uluna"
            }]
        }, {
            "type": "transfer",
            "attributes": [{
                "key": "recipient",
                "value": "terra1jv65s3grqf6v6jl3dp4t6c9t9rk99cd8pm7utl"
            }, {
                "key": "sender",
                "value": tokenFactoryWalletAddr
            }, {
                "key": "amount",
                "value": "10000000uluna"
            }]
        }, {
            "type": "message",
            "attributes": [{
                "key": "sender",
                "value": tokenFactoryWalletAddr
            }]
        }, {
            "type": "create_denom",
            "attributes": [{
                "key": "creator",
                "value": tokenFactoryWalletAddr
            }, {
                "key": "new_token_denom",
                "value": factoryDenom
            }]
        }]);
    })

    // Mint tokens to the minter address
    // and assert the logs are correctly formatted.
    describe("After creating the token", () => {
        test('Must mint some tokens', async () => {
            let tx = await wallet.createAndSignTx({
                msgs: [
                    new MsgMint(
                        tokenFactoryWalletAddr,
                        Coin.fromString("1000000000" + factoryDenom)
                    ),
                ],
                chainID: "test-1",
            });
            let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            expect(txResult.logs[0].events).toStrictEqual([{
                "type": "message",
                "attributes": [{
                    "key": "action",
                    "value": "/osmosis.tokenfactory.v1beta1.MsgMint"
                }, {
                    "key": "sender",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "module",
                    "value": "tokenfactory"
                }]
            },
            {
                "type": "coin_received",
                "attributes": [{
                    "key": "receiver",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "1000000000" + factoryDenom
                }]
            },
            {
                "type": "coinbase",
                "attributes": [{
                    "key": "minter",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "1000000000" + factoryDenom
                }]
            },
            {
                "type": "coin_spent",
                "attributes": [{
                    "key": "spender",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "1000000000" + factoryDenom
                }]
            },
            {
                "type": "coin_received",
                "attributes": [{
                    "key": "receiver",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "amount",
                    "value": "1000000000" + factoryDenom
                }]
            },
            {
                "type": "transfer",
                "attributes": [{
                    "key": "recipient",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "sender",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "1000000000" + factoryDenom
                }]
            },
            {
                "type": "message",
                "attributes": [{
                    "key": "sender",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }]
            },
            {
                "type": "tf_mint",
                "attributes": [{
                    "key": "mint_to_address",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "amount",
                    "value": "1000000000" + factoryDenom
                }]
            }]);
        });
    })

    // Burn some tokens from the minter account 
    // and asser the logs are correctly formatted.
    describe("After minting the tokens", () => {
        test('Must burn some tokens', async () => {
            let tx = await wallet.createAndSignTx({
                msgs: [
                    new MsgBurn(
                        tokenFactoryWalletAddr,
                        Coin.fromString("500000000" + factoryDenom)
                    ),
                ],
                chainID: "test-1",
                fee: new Fee(100_000, new Coins({ uluna: 100_000 })),
            });
            let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            expect(txResult.logs[0].events).toStrictEqual([{
                "type": "message",
                "attributes": [{
                    "key": "action",
                    "value": "/osmosis.tokenfactory.v1beta1.MsgBurn"
                }, {
                    "key": "sender",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "module",
                    "value": "tokenfactory"
                }]
            }, {
                "type": "coin_spent",
                "attributes": [{
                    "key": "spender",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "amount",
                    "value": "500000000" + factoryDenom
                }]
            }, {
                "type": "coin_received",
                "attributes": [{
                    "key": "receiver",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "500000000" + factoryDenom
                }]
            }, {
                "type": "transfer",
                "attributes": [{
                    "key": "recipient",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "sender",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "amount",
                    "value": "500000000" + factoryDenom
                }]
            }, {
                "type": "message",
                "attributes": [{
                    "key": "sender",
                    "value": tokenFactoryWalletAddr
                }
                ]
            }, {
                "type": "coin_spent",
                "attributes": [{
                    "key": "spender",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "500000000" + factoryDenom
                }]
            }, {
                "type": "burn",
                "attributes": [{
                    "key": "burner",
                    "value": "terra19ejy8n9qsectrf4semdp9cpknflld0j6my8d0p"
                }, {
                    "key": "amount",
                    "value": "500000000" + factoryDenom
                }]
            }, {
                "type": "tf_burn",
                "attributes": [{
                    "key": "burn_from_address",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "amount",
                    "value": "500000000" + factoryDenom
                }]
            }])
        });
    })


    describe("Use before send hooks", () => {
        // Validate update params using gov proposal
        describe("Add hooks to the no100 contract", () => {

            test('Must update params using gov proposal', async () => {
                // let blockHeight = (await LCD.chain1.tendermint.blockInfo("test-1")).block.header.height;
                let tx = await val1Wallet.createAndSignTx({
                    msgs: [new MsgSubmitProposal([
                        new MsgUpdateParamsTokenFactory(
                            govAddress,
                            Coins.fromString("512000uluna"),
                            1000000,
                            [
                                { "codeId": codeId, "denomCreator": tokenFactoryWalletAddr }
                            ]
                        )
                    ],
                        Coins.fromString("1000000000uluna"),
                        val1WalletAddress,
                        "metadata",
                        "title",
                        "summary"
                    )],
                    chainID: "test-1",
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();

                // Check that the proposal was created successfully
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.code).toBe(0);

                // Get the proposal id and validate exists
                let proposalId = Number(txResult.logs[0].eventsByType.submit_proposal.proposal_id[0]);
                expect(proposalId)

                // Vote for the proposal
                tx = await val1Wallet.createAndSignTx({
                    msgs: [new MsgVote(
                        proposalId,
                        val1WalletAddress,
                        VoteOption.VOTE_OPTION_YES
                    )],
                    fee: new Fee(100_000, "100000uluna"),
                    chainID: "test-1",
                });
                result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await votingPeriod();
                txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1")
                expect(txResult.code).toBe(0);

                const params = await LCD.chain1.tokenfactory.params("test-1");
                expect(params).toStrictEqual({
                    params: {
                        denom_creation_fee: [{
                            amount: "512000",
                            denom: "uluna"
                        }],
                        denom_creation_gas_consume: '1000000',
                        whitelisted_hooks: [{
                            code_id: "" + codeId,
                            denom_creator: tokenFactoryWalletAddr
                        }]
                    }
                });
            })

            test("Must register the hooks to the no100 contract", async () => {
                let tx = await wallet.createAndSignTx({
                    msgs: [
                        new MsgSetBeforeSendHook(
                            tokenFactoryWalletAddr,
                            factoryDenom as string,
                            contractAddress,
                        ),
                    ],
                    fee: new Fee(100_000, new Coins({ uluna: 100_000 })),
                    chainID: "test-1",
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.logs[0].events).toStrictEqual([{
                    "type": "message",
                    "attributes": [{
                        "key": "action",
                        "value": "/osmosis.tokenfactory.v1beta1.MsgSetBeforeSendHook"
                    }, {
                        "key": "sender",
                        "value": tokenFactoryWalletAddr
                    }, {
                        "key": "module",
                        "value": "tokenfactory"
                    }]
                }, {
                    "type": "set_before_send_hook",
                    "attributes": [{
                        "key": "denom",
                        "value": factoryDenom
                    }, {
                        "key": "before_send_hook_address",
                        "value": contractAddress
                    }]
                }]);
            });
        });


        // This test proves that the wasm contract 
        // is being executed on the sudo before send 
        // hook, one test allows transaction and the 
        // other one blocks the transaction.
        describe("Must send tokens and be intercepted by beforesendhooks", () => {
            test("1 token successfuly", async () => {
                let tx = await wallet.createAndSignTx({
                    msgs: [
                        new MsgSend(
                            tokenFactoryWalletAddr,
                            randomAccountAddr,
                            Coins.fromString("1" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.logs[0].events)
                    .toStrictEqual([{
                        "type": "message",
                        "attributes": [{
                            "key": "action",
                            "value": "/cosmos.bank.v1beta1.MsgSend"
                        }, {
                            "key": "sender",
                            "value": tokenFactoryWalletAddr
                        }, {
                            "key": "module",
                            "value": "bank"
                        }]
                    }, {
                        "type": "coin_spent",
                        "attributes": [{
                            "key": "spender",
                            "value": tokenFactoryWalletAddr
                        }, {
                            "key": "amount",
                            "value": "1" + factoryDenom
                        }]
                    }, {
                        "type": "coin_received",
                        "attributes": [{
                            "key": "receiver",
                            "value": randomAccountAddr
                        }, {
                            "key": "amount",
                            "value": "1" + factoryDenom
                        }]
                    }, {
                        "type": "transfer",
                        "attributes": [{
                            "key": "recipient",
                            "value": randomAccountAddr
                        }, {
                            "key": "sender",
                            "value": tokenFactoryWalletAddr
                        }, {
                            "key": "amount",
                            "value": "1" + factoryDenom
                        }]
                    }, {
                        "type": "message",
                        "attributes": [{
                            "key": "sender",
                            "value": tokenFactoryWalletAddr
                        }]
                    }]);
            });

            test("100 token blocked by the smart contract before send", async () => {
                let tx = await wallet.createAndSignTx({
                    msgs: [
                        new MsgSend(
                            tokenFactoryWalletAddr,
                            randomAccountAddr,
                            Coins.fromString("100" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                    fee: new Fee(2000_000, new Coins({ uluna: 100_000 })),
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.raw_log)
                    .toStrictEqual(`failed to execute message; message index: 0: failed to call before send hook for denom ${factoryDenom}: Custom Error val: \"Invalid Send Amount\": execute wasm contract failed`);
            });

            test("100 token blocked by the smart contract on burn", async () => {
                let tx = await wallet.createAndSignTx({
                    msgs: [
                        new MsgBurn(
                            tokenFactoryWalletAddr,
                            Coin.fromString("100" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                    fee: new Fee(2000_000, new Coins({ uluna: 100_000 })),
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.raw_log)
                    .toStrictEqual(`failed to execute message; message index: 0: failed to call before send hook for denom ${factoryDenom}: Custom Error val: \"Invalid Send Amount\": execute wasm contract failed`);
            });

            test("100 token blocked by the smart contract on mint", async () => {
                let tx = await wallet.createAndSignTx({
                    msgs: [
                        new MsgMint(
                            tokenFactoryWalletAddr,
                            Coin.fromString("100" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                    fee: new Fee(2000_000, new Coins({ uluna: 100_000 })),
                });
                let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.raw_log)
                    .toStrictEqual(`failed to execute message; message index: 0: failed to call before send hook for denom ${factoryDenom}: Custom Error val: \"Invalid Send Amount\": execute wasm contract failed`);
            });
        });
    })


    // Change the token admin to a random account 
    // to validate that the functionality works and 
    // assert the logs are correctly formatted.
    describe("After all operations", () => {
        test("Must change the admin of the denom", async () => {
            let tx = await wallet.createAndSignTx({
                msgs: [
                    new MsgChangeAdmin(
                        tokenFactoryWalletAddr,
                        randomAccountAddr,
                        factoryDenom as string,
                    ),
                ],
                fee: new Fee(100_000, new Coins({ uluna: 100_000 })),
                chainID: "test-1",
            });
            let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            expect(txResult.logs[0].events).toStrictEqual([{
                "type": "message",
                "attributes": [{
                    "key": "action",
                    "value": "/osmosis.tokenfactory.v1beta1.MsgChangeAdmin"
                }, {
                    "key": "sender",
                    "value": tokenFactoryWalletAddr
                }, {
                    "key": "module",
                    "value": "tokenfactory"
                }]
            },
            {
                "type": "change_admin",
                "attributes": [{
                    "key": "denom",
                    "value": factoryDenom
                }, {
                    "key": "new_admin",
                    "value": randomAccountAddr
                }]
            }]);
        });

        test("Must query the new admin of the denom", async () => {
            const res = await LCD.chain1.tokenfactory.authorityMetadata("test-1", encodeURIComponent(encodeURIComponent(factoryDenom as string)));

            expect(res)
                .toStrictEqual({
                    "authority_metadata": {
                        "admin": randomAccountAddr
                    }
                })
        })

        test("Must query the before send hook", async () => {
            const res = await LCD.chain1.tokenfactory.beforeSendHook("test-1", encodeURIComponent(encodeURIComponent(factoryDenom as string)));

            expect(res)
                .toStrictEqual({ "cosmwasm_address": contractAddress })
        })

        test("Must query the before send hook", async () => {
            const res = await LCD.chain1.tokenfactory.denomsFromCreator(tokenFactoryWalletAddr);
            expect(res.denoms.length).toBeGreaterThanOrEqual(1);
        })
    })

    describe("Query using CosmWasm", () => {
        beforeAll(async () => {
            // Deploy alliance query contract
            let tx = await wallet.createAndSignTx({
                msgs: [
                    new MsgStoreCode(tokenFactoryWalletAddr, fs.readFileSync(path.join(__dirname, "/../../contracts/custom_queries.wasm")).toString("base64")),
                ],
                chainID: "test-1",
            });

            let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();

            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            let customQueryCodeId = Number(txResult.logs[0].events[1].attributes[1].value);
            expect(customQueryCodeId).toBeDefined();

            // Instantiate alliance query contract
            tx = await wallet.createAndSignTx({
                msgs: [new MsgInstantiateContract(
                    tokenFactoryWalletAddr,
                    tokenFactoryWalletAddr,
                    customQueryCodeId,
                    {},
                    undefined,
                    "Alliance query contract" + Math.random(),
                )],
                chainID: "test-1",
            });
            result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
            await blockInclusion();

            txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            customQueryContractAddress = txResult.logs[0].events[1].attributes[0].value;
            expect(customQueryContractAddress).toBeDefined();
        })
        test("Must query token data using contract", async () => {
            let res = await LCD.chain1.wasm.contractQuery(customQueryContractAddress, {
                token: {
                    admin: {
                        denom: factoryDenom
                    }
                }
            }) as any;
            expect(res).toStrictEqual({
                admin: randomAccountAddr
            });
        })
    })
});