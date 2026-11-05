import { Coin, Coins, MnemonicKey, MsgBurn, MsgChangeAdmin, MsgCreateDenom, MsgInstantiateContract, MsgMint, MsgStoreCode, MsgSetBeforeSendHook, MsgSend, MsgSubmitProposal, MsgUpdateParamsTokenFactory, MsgVote } from "@terra-money/feather.js";
import { getMnemonics, getLCDClient, blockInclusion, votingPeriod, getValueByIndexAndTypeAndKey, signAndBroadcastTx } from "../../helpers";

import fs from "fs";
import path from 'path';

describe("TokenFactory Module", () => {
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

    const expectEventAttribute = (events: any[], eventType: string, key: string, value: string) => {
        expect(events.some((event: any) =>
            event.type === eventType &&
            event.attributes.some((attr: any) => attr.key === key && attr.value === value)
        )).toBeTruthy();
    };
    const getProposalId = (events: any[]) => Number(
        events
            .flatMap((event: any) => event.attributes)
            .find((attr: any) => attr.key === "proposal_id")?.value
    );

    // Read the no100 contract, store on chain, 
    // instantiate to be used in the following tests
    // and finally save the contract address.
    beforeAll(async () => {
        let result = await signAndBroadcastTx(wallet, {
            msgs: [new MsgStoreCode(
                tokenFactoryWalletAddr,
                fs.readFileSync(path.join(__dirname, "/../../contracts/no100.wasm")).toString("base64"),
            )],
            chainID: "test-1",
        });

        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        codeId = Number(getValueByIndexAndTypeAndKey(txResult.events, 0, "store_code", "code_id"));
        expect(codeId).toBeDefined();

        const msgInstantiateContract = new MsgInstantiateContract(
            tokenFactoryWalletAddr,
            tokenFactoryWalletAddr,
            codeId,
            {},
            Coins.fromString("1uluna"),
            "no100 contract " + Math.random(),
        );

        result = await signAndBroadcastTx(wallet, {
            msgs: [msgInstantiateContract],
            chainID: "test-1",
        });
        await blockInclusion();
        txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        contractAddress = getValueByIndexAndTypeAndKey(txResult.events, 0, "instantiate", "_contract_address");
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
        let result = await signAndBroadcastTx(wallet, {
            msgs: [
                new MsgCreateDenom(
                    tokenFactoryWalletAddr,
                    subdenom,
                ),
            ],
            chainID: "test-1",
        });
        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        factoryDenom = getValueByIndexAndTypeAndKey(txResult.events, 0, "create_denom", "new_token_denom") as string
        expect(factoryDenom).toBeDefined();
        expectEventAttribute(txResult.events, "message", "action", "/osmosis.tokenfactory.v1beta1.MsgCreateDenom");
        expectEventAttribute(txResult.events, "create_denom", "creator", tokenFactoryWalletAddr);
        expectEventAttribute(txResult.events, "create_denom", "new_token_denom", factoryDenom);
    })

    // Mint tokens to the minter address
    // and assert the logs are correctly formatted.
    describe("After creating the token", () => {
        test('Must mint some tokens', async () => {
            let result = await signAndBroadcastTx(wallet, {
                msgs: [
                    new MsgMint(
                        tokenFactoryWalletAddr,
                        Coin.fromString("1000000000" + factoryDenom)
                    ),
                ],
                chainID: "test-1",
            });
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            expectEventAttribute(txResult.events, "message", "action", "/osmosis.tokenfactory.v1beta1.MsgMint");
            expectEventAttribute(txResult.events, "tf_mint", "mint_to_address", tokenFactoryWalletAddr);
            expectEventAttribute(txResult.events, "tf_mint", "amount", "1000000000" + factoryDenom);
        });
    })

    // Burn some tokens from the minter account 
    // and asser the logs are correctly formatted.
    describe("After minting the tokens", () => {
        test('Must burn some tokens', async () => {
            let result = await signAndBroadcastTx(wallet, {
                msgs: [
                    new MsgBurn(
                        tokenFactoryWalletAddr,
                        Coin.fromString("500000000" + factoryDenom)
                    ),
                ],
                chainID: "test-1",
                fee: {
                    amount: [{ denom: "uluna", amount: "100000" }],
                    gas: "200000",
                },
            });
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            expect(txResult.code).toBe(0);
            expectEventAttribute(txResult.events, "message", "action", "/osmosis.tokenfactory.v1beta1.MsgBurn");
            expectEventAttribute(txResult.events, "tf_burn", "burn_from_address", tokenFactoryWalletAddr);
            expectEventAttribute(txResult.events, "tf_burn", "amount", "500000000" + factoryDenom);
        });
    })


    describe("Use before send hooks", () => {
        // Validate update params using gov proposal
        describe("Add hooks to the no100 contract", () => {

            test('Must update params using gov proposal', async () => {
                // let blockHeight = (await LCD.chain1.tendermint.blockInfo("test-1")).block.header.height;
                let result = await signAndBroadcastTx(val1Wallet, {
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
                await blockInclusion();

                // Check that the proposal was created successfully
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.code).toBe(0);

                // Get the proposal id and validate exists
                let proposalId = getProposalId(txResult.events);
                expect(proposalId)

                // Vote for the proposal
                result = await signAndBroadcastTx(val1Wallet, {
                    msgs: [new MsgVote(
                        proposalId,
                        val1WalletAddress,
                        1 // Yes
                    )],
                    fee: {
                        amount: [{ denom: "uluna", amount: "100000" }],
                        gas: "100000",
                    },
                    chainID: "test-1",
                });
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
                let result = await signAndBroadcastTx(wallet, {
                    msgs: [
                        new MsgSetBeforeSendHook(
                            tokenFactoryWalletAddr,
                            factoryDenom as string,
                            contractAddress,
                        ),
                    ],
                    fee: {
                        amount: [{ denom: "uluna", amount: "100000" }],
                        gas: "100000",
                    },
                    chainID: "test-1",
                });
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expectEventAttribute(txResult.events, "message", "action", "/osmosis.tokenfactory.v1beta1.MsgSetBeforeSendHook");
                expectEventAttribute(txResult.events, "set_before_send_hook", "denom", factoryDenom as string);
                expectEventAttribute(txResult.events, "set_before_send_hook", "before_send_hook_address", contractAddress);
            });
        });


        // This test proves that the wasm contract 
        // is being executed on the sudo before send 
        // hook, one test allows transaction and the 
        // other one blocks the transaction.
        describe("Must send tokens and be intercepted by beforesendhooks", () => {
            test("1 token successfuly", async () => {
                let result = await signAndBroadcastTx(wallet, {
                    msgs: [
                        new MsgSend(
                            tokenFactoryWalletAddr,
                            randomAccountAddr,
                            Coins.fromString("1" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                });
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expectEventAttribute(txResult.events, "message", "action", "/cosmos.bank.v1beta1.MsgSend");
                expectEventAttribute(txResult.events, "transfer", "recipient", randomAccountAddr);
                expectEventAttribute(txResult.events, "transfer", "amount", "1" + factoryDenom);
            });

            test("100 token blocked by the smart contract before send", async () => {
                let result = await signAndBroadcastTx(wallet, {
                    msgs: [
                        new MsgSend(
                            tokenFactoryWalletAddr,
                            randomAccountAddr,
                            Coins.fromString("100" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                    fee: {
                        amount: [{ denom: "uluna", amount: "100000" }],
                        gas: "2000000",
                    },
                });
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.raw_log)
                    .toStrictEqual(`failed to execute message; message index: 0: failed to call before send hook for denom ${factoryDenom}: Custom Error val: \"Invalid Send Amount\": execute wasm contract failed`);
            });

            test("100 token blocked by the smart contract on burn", async () => {
                let result = await signAndBroadcastTx(wallet, {
                    msgs: [
                        new MsgBurn(
                            tokenFactoryWalletAddr,
                            Coin.fromString("100" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                    fee: {
                        amount: [{ denom: "uluna", amount: "100000" }],
                        gas: "2000000",
                    },
                });
                await blockInclusion();
                let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
                expect(txResult.raw_log)
                    .toStrictEqual(`failed to execute message; message index: 0: failed to call before send hook for denom ${factoryDenom}: Custom Error val: \"Invalid Send Amount\": execute wasm contract failed`);
            });

            test("100 token blocked by the smart contract on mint", async () => {
                let result = await signAndBroadcastTx(wallet, {
                    msgs: [
                        new MsgMint(
                            tokenFactoryWalletAddr,
                            Coin.fromString("100" + factoryDenom),
                        ),
                    ],
                    chainID: "test-1",
                    fee: {
                        amount: [{ denom: "uluna", amount: "100000" }],
                        gas: "2000000",
                    },
                });
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
            let result = await signAndBroadcastTx(wallet, {
                msgs: [
                    new MsgChangeAdmin(
                        tokenFactoryWalletAddr,
                        randomAccountAddr,
                        factoryDenom as string,
                    ),
                ],
                fee: {
                    amount: [{ denom: "uluna", amount: "100000" }],
                    gas: "100000",
                },
                chainID: "test-1",
            });
            await blockInclusion();
            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            expectEventAttribute(txResult.events, "message", "action", "/osmosis.tokenfactory.v1beta1.MsgChangeAdmin");
            expectEventAttribute(txResult.events, "change_admin", "denom", factoryDenom as string);
            expectEventAttribute(txResult.events, "change_admin", "new_admin", randomAccountAddr);
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
            let result = await signAndBroadcastTx(wallet, {
                msgs: [
                    new MsgStoreCode(tokenFactoryWalletAddr, fs.readFileSync(path.join(__dirname, "/../../contracts/custom_queries.wasm")).toString("base64")),
                ],
                chainID: "test-1",
            });

            await blockInclusion();

            let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            let customQueryCodeId = Number(getValueByIndexAndTypeAndKey(txResult.events, 0, "store_code", "code_id"));
            expect(customQueryCodeId).toBeDefined();

            // Instantiate alliance query contract
            result = await signAndBroadcastTx(wallet, {
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
            await blockInclusion();

            txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
            customQueryContractAddress = getValueByIndexAndTypeAndKey(txResult.events, 0, "instantiate", "_contract_address");
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
