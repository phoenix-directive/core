import { getLCDClient, getMnemonics, blockInclusion, votingPeriod, getValueByIndexAndTypeAndKey, } from "../../helpers";
import {
    Coin,
    MsgCreateAlliance,
    Coins,
    MsgVote,
    Fee,
    MsgAllianceDelegate,
    MsgClaimDelegationRewards,
    MsgAllianceUndelegate,
    MsgDeleteAlliance,
    MsgSubmitProposal,
    MsgStoreCode, MsgInstantiateContract,
    MsgCreateDenom,
    MsgMint
} from "@terra-money/feather.js";
import { VoteOption } from "@terra-money/terra.proto/cosmos/gov/v1beta1/gov";
import fs from "fs";
import path from "path";

describe("Alliance Module (https://github.com/terra-money/alliance/tree/release/v0.3.x) ", () => {
    // Prepare environment clients, accounts and wallets
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    // const chain1Wallet = LCD.chain1.wallet(accounts.allianceMnemonic);
    const allianceWallet2 = LCD.chain2.wallet(accounts.allianceMnemonic);
    const val2Wallet = LCD.chain2.wallet(accounts.val2);
    const val2WalletAddress = val2Wallet.key.accAddress("terra");
    const val2Address = val2Wallet.key.valAddress("terra");
    const allianceAccountAddress = accounts.allianceMnemonic.accAddress("terra");
    // This will be populated in the "Must create an alliance"
    let allianceCoin = Coin.fromString("1uluna");
    let subdenom = Math.random().toString(36).substring(7);
    let allianceQueryCodeId: number;
    let allianceQueryContract: string;

    // Send uluna from chain-1 to chain-2 using 
    // the same wallet on both chains and start
    // an Alliance creation process
    beforeAll(async () => {
        // let blockHeight = (await LCD.chain1.tendermint.blockInfo("test-1")).block.header.height;
        // let tx = await chain1Wallet.createAndSignTx({
        //     msgs: [new MsgTransfer(
        //         "transfer",
        //         "channel-0",
        //         Coin.fromString("100000000uluna"),
        //         allianceAccountAddress,
        //         allianceAccountAddress,
        //         new Height(2, parseInt(blockHeight) + 100),
        //         undefined,
        //         ""
        //     )],
        //     chainID: "test-1",
        // });

        // let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        // await blockInclusion();
        // let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        // expect(txResult).toBeDefined();

        // // Check during 5 blocks for the receival 
        // // of the IBC coin on chain-2
        // for (let i = 0; i <= 5; i++) {
        //     await blockInclusion();
        //     let _ibcCoin = (await LCD.chain2.bank.balance(allianceAccountAddress))[0].find(c => c.denom.startsWith("ibc/"));
        //     if (_ibcCoin) {
        //         expect(_ibcCoin.denom.startsWith("ibc/")).toBeTruthy();
        //         break;
        //     }
        // }
        let tx = await val2Wallet.createAndSignTx({
            msgs: [
                new MsgCreateDenom(
                    val2WalletAddress,
                    subdenom,
                ),
            ],
            chainID: "test-2",
        });

        let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
        await blockInclusion();
        await blockInclusion();
        let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        allianceCoin = new Coin(getValueByIndexAndTypeAndKey(txResult.events, 0, "create_denom", "new_token_denom"), 1);

        tx = await val2Wallet.createAndSignTx({
            msgs: [
                new MsgMint(
                    val2WalletAddress,
                    new Coin(allianceCoin.denom, 1000000000000000000),
                ),
            ],
            chainID: "test-2",
        });
        result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
        await blockInclusion();

        const balances = await LCD.chain2.bank.balance(val2WalletAddress);
        expect(balances[0].find(b => b.denom === allianceCoin.denom)?.amount.toString()).toBe("1000000000000000000");
    });

    test('Must contain the expected module params', async () => {
        // Query Alliance module params
        const moduleParams = await LCD.chain2.alliance.params("test-2");

        // Validate that the params were set correctly on genesis
        expect(moduleParams.takeRateClaimInterval)
            .toBe("300s");
        expect(moduleParams.rewardDelayTime)
            .toBe("0s");
    });

    test('Must create an alliance', async () => {
        const msgProposal = new MsgSubmitProposal(
            [new MsgCreateAlliance(
                "terra10d07y265gmmuvt4z0w9aw880jnsr700juxf95n",
                allianceCoin.denom,
                "100000000000000000",
                "0",
                "1000000000000000000",
                undefined,
                {
                    "min": "100000000000000000",
                    "max": "100000000000000000"
                })],
            Coins.fromString("1000000000uluna"),
            val2WalletAddress,
            "metadata",
            "title",
            "summary"
        );
        // Create an alliance proposal sign and submit on chain-2
        let tx = await val2Wallet.createAndSignTx({
            msgs: [msgProposal],
            chainID: "test-2",
        });
        let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
        await blockInclusion();

        // Check that the proposal was created successfully
        let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        expect(txResult.code).toBe(0);

        // Get the proposal id and validate exists
        const proposalId = getValueByIndexAndTypeAndKey(txResult.events, 0, "submit_proposal", "proposal_id");
        expect(proposalId)

        // Vote for the proposal
        tx = await val2Wallet.createAndSignTx({
            msgs: [new MsgVote(
                proposalId,
                val2WalletAddress,
                VoteOption.VOTE_OPTION_YES
            )],
            fee: new Fee(100_000, "100000uluna"),
            chainID: "test-2",
        });
        result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
        await blockInclusion();
        await votingPeriod();
        txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2")
        expect(txResult.code).toBe(0);

        const res = await LCD.chain2.alliance.queryAlliance("test-2", allianceCoin.denom);

        expect(res).toBeDefined();
        expect(res.denom).toBe(allianceCoin.denom);
        expect(res.rewardWeight.toString()).toBe("0.100000000000000000");
        expect(res.takeRate.toString()).toBe("0.000000000000000000");
        expect(res.rewardWeightRange?.min.toString()).toBe("0.100000000000000000")
        expect(res.rewardWeightRange?.max.toString()).toBe("0.100000000000000000")
        expect(res.isInitialized).toBeTruthy();
    });

    describe("After Alliance has been created", () => {
        test('Must delegate to the alliance', async () => {
            let tx = await val2Wallet.createAndSignTx({
                msgs: [
                    new MsgAllianceDelegate(
                        val2WalletAddress,
                        val2Address,
                        new Coin(allianceCoin.denom, 1000),
                    )
                ],
                chainID: "test-2",
            });
            let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
            await blockInclusion();

            // Check that the proposal was created successfully
            let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
            expect(txResult.code).toBe(0);

            // Validate the delegation event
            expect(getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.DelegateAllianceEvent", "allianceSender"))
                .toStrictEqual(`"${val2WalletAddress}"`)
            expect(getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.DelegateAllianceEvent", "coin"))
                .toStrictEqual(`{\"denom\":\"${allianceCoin.denom}\",\"amount\":\"1000\"}`)
            expect(getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.DelegateAllianceEvent", "newShares"))
                .toStrictEqual(`"1000.000000000000000000"`)
            expect(getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.DelegateAllianceEvent", "validator"))
                .toStrictEqual(`"${val2Address}"`)
        });

        test('Must query one alliance validators', async () => {
            const res = await LCD.chain2.alliance.queryAllianceValidators("test-2", val2Address);
            expect(res.validators[0].totalDelegationShares.find(c => c.denom === allianceCoin.denom)?.amount.toString()).toBe("1000.000000000000000000");
            expect(res.validators[0].validatorShares.find(c => c.denom === allianceCoin.denom)?.amount.toString()).toBe("1000.000000000000000000");
            expect(res.validators[0].totalStaked.find(c => c.denom === allianceCoin.denom)?.amount.toString()).toBe("1000.000000000000000000");
        });

        test('Must query all alliance validators', async () => {
            const res = await LCD.chain2.alliance.queryAllianceValidators("test-2");
            const validatorRes = res.validators.find(v => v.validatorAddr === val2Address);
            expect(validatorRes).toBeDefined();
            expect(validatorRes?.totalDelegationShares.find(c => c.denom === allianceCoin.denom)?.amount.toString()).toBe("1000.000000000000000000");
            expect(validatorRes?.validatorShares.find(c => c.denom === allianceCoin.denom)?.amount.toString()).toBe("1000.000000000000000000");
            expect(validatorRes?.totalStaked.find(c => c.denom === allianceCoin.denom)?.amount.toString()).toBe("1000.000000000000000000");
        });

        describe("Alliance wasm queries", () => {
            beforeAll(async () => {

                // Deploy query contract
                let tx = await allianceWallet2.createAndSignTx({
                    msgs: [
                        new MsgStoreCode(allianceAccountAddress, fs.readFileSync(path.join(__dirname, "/../../contracts/custom_queries.wasm")).toString("base64")),
                    ],
                    chainID: "test-2",
                });

                let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
                await blockInclusion();

                let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;

                allianceQueryCodeId = Number(getValueByIndexAndTypeAndKey(txResult.events, 0, "store_code", "code_id"));
                expect(allianceQueryCodeId).toBeDefined();

                // Instantiate query contract
                tx = await allianceWallet2.createAndSignTx({
                    msgs: [new MsgInstantiateContract(
                        allianceAccountAddress,
                        allianceAccountAddress,
                        allianceQueryCodeId,
                        {},
                        undefined,
                        "Alliance query contract" + Math.random(),
                    )],
                    chainID: "test-2",
                });
                result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
                await blockInclusion();

                txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
                allianceQueryContract = getValueByIndexAndTypeAndKey(txResult.events, 0,"instantiate", "_contract_address");
                expect(allianceQueryContract).toBeDefined();
            })

            test("Must be able to query alliance state in CosmWasm", async () => {

                let res = await LCD.chain2.wasm.contractQuery(allianceQueryContract, {
                    alliance: {
                        denom: allianceCoin.denom
                    }
                }) as any;
                expect(res.denom).toEqual(allianceCoin.denom);

                res = await LCD.chain2.wasm.contractQuery(allianceQueryContract, {
                    delegation: {
                        denom: allianceCoin.denom,
                        validator: val2Address,
                        delegator: val2WalletAddress,
                    }
                }) as any;
                expect(res).toStrictEqual({
                    delegator: val2WalletAddress,
                    validator: val2Address,
                    denom: allianceCoin.denom,
                    amount: "1000",
                });

                await blockInclusion();
                res = await LCD.chain2.wasm.contractQuery(allianceQueryContract, {
                    delegation_rewards: {
                        denom: allianceCoin.denom,
                        validator: val2Address,
                        delegator: val2WalletAddress,
                    }
                }) as any;
                expect(res.rewards.length).toEqual(1);
            })
        })

        describe("After delegation", () => {
            test("Must claim rewards from the alliance", async () => {
                let tx = await val2Wallet.createAndSignTx({
                    msgs: [
                        new MsgClaimDelegationRewards(
                            val2WalletAddress,
                            val2Address,
                            allianceCoin.denom,
                        ),
                    ],
                    fee: new Fee(1000_000, "100000uluna"),
                    chainID: "test-2",
                });
                let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
                await blockInclusion();

                // Check that the proposal was created successfully
                let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
                expect(txResult.code).toBe(0);

                // Validate the delegation event
                const allianceSender = getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.ClaimAllianceRewardsEvent", "allianceSender");
                expect(allianceSender)
                    .toStrictEqual(`"${val2WalletAddress}"`)
                const validator = getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.ClaimAllianceRewardsEvent", "validator");
                expect(validator)
                    .toStrictEqual(`"${val2Address}"`)
            })

            test("Must undelegate from the alliance", async () => {
                await blockInclusion();
                let tx = await val2Wallet.createAndSignTx({
                    msgs: [
                        new MsgAllianceUndelegate(
                            val2WalletAddress,
                            val2Address,
                            new Coin(allianceCoin.denom, 1000),
                        ),
                    ],
                    fee: new Fee(1000_000, "100000uluna"),
                    chainID: "test-2",
                });
                let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
                await blockInclusion();
                await blockInclusion();

                // Check that the proposal was created successfully
                let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
                expect(txResult.code).toBe(0);

                // Validate the delegation event
                const allianceSender = getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.UndelegateAllianceEvent", "allianceSender");
                expect(allianceSender)
                    .toStrictEqual(`"${val2WalletAddress}"`)
                const validator = getValueByIndexAndTypeAndKey(txResult.events, 0, "alliance.alliance.UndelegateAllianceEvent", "validator");
                expect(validator)
                    .toStrictEqual(`"${val2Address}"`)
            })
        })
    })

    describe("After interacting with the Alliance", () => {
        test('Must removed the alliance using gov', async () => {
            // Query the alliance and check if it exists
            const oldRes = await LCD.chain2.alliance.queryAlliances("test-2");
            const oldCount = oldRes.alliances.length;

            const msgProposal = new MsgSubmitProposal(
                [new MsgDeleteAlliance(
                    "terra10d07y265gmmuvt4z0w9aw880jnsr700juxf95n",
                    allianceCoin.denom,
                )],
                Coins.fromString("1000000000uluna"),
                val2WalletAddress,
                "metadata",
                "title",
                "summary"
            );
            // Create a delete alliance proposal sign and submit on chain-2
            let tx = await val2Wallet.createAndSignTx({
                msgs: [msgProposal],
                chainID: "test-2",
            });
            let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
            await blockInclusion();

            // Check that the proposal was created successfully
            let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
            expect(txResult.code).toBe(0);

            // Get the proposal id and validate exists
            let proposalId = Number(getValueByIndexAndTypeAndKey(txResult.events, 0, "proposal_deposit", "proposal_id"));
            expect(proposalId)

            // Vote for the proposal
            tx = await val2Wallet.createAndSignTx({
                msgs: [new MsgVote(
                    proposalId,
                    val2WalletAddress,
                    VoteOption.VOTE_OPTION_YES
                )],
                fee: new Fee(100_000, "100000uluna"),
                chainID: "test-2",
            });
            result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
            await votingPeriod();
            txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2")
            expect(txResult.code).toBe(0);

            // Query the alliance and check if it exists
            const res = await LCD.chain2.alliance.queryAlliances("test-2");
            expect(res.alliances.length).toBe(oldCount - 1);
        });
    })
});