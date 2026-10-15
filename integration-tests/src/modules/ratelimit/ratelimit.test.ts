import { Coin, Coins, MsgSubmitProposal, MsgTransfer, MsgVote } from "@terra-money/feather.js";
import { VoteOption } from "@terra-money/terra.proto/cosmos/gov/v1beta1/gov";
import moment from "moment";
import {
    blockInclusion,
    getLCDClient,
    getMnemonics,
    getValueByIndexAndTypeAndKey,
    ibcTransfer,
    msgAddRateLimit,
    msgRemoveRateLimit,
    pollUntil,
    queryAllRateLimits,
    queryRateLimit,
    signAndBroadcastTx,
    votingPeriod,
} from "../../helpers";
import type { RateLimit } from "../../helpers/ratelimit";

/**
 * Rate limits are configured on **chain-2** for its native `uluna` over
 * `channel-0`, so the outflow is exercised by transferring chain-2 luna to
 * chain-1 and the inflow by sending the resulting voucher back.
 *
 * The quota is a percentage of the channel value, which the module snapshots
 * from the total supply when the window opens. Inflation keeps changing that
 * supply, so the test reads the recorded channel value back and derives every
 * transfer amount from it instead of hardcoding one.
 */
describe("Ratelimit Module (github.com/cosmos/ibc-apps/modules/rate-limiting)", () => {
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    const val2Wallet = LCD.chain2.wallet(accounts.val2);
    const val2WalletAddress = accounts.val2.accAddress("terra");
    // wallet14, funded with 1_000_000_000_000uluna on both chains and unused by
    // the other suites, so its balance and sequence cannot be disturbed.
    const chain1Wallet = LCD.chain1.wallet(accounts.mnemonic5);
    const chain2Wallet = LCD.chain2.wallet(accounts.mnemonic5);
    const walletAddress = accounts.mnemonic5.accAddress("terra");

    const CHANNEL = "channel-0";
    const DENOM = "uluna";
    // chain-2 uluna as it is denominated on chain-1 after crossing channel-0.
    const IBC_DENOM_ON_CHAIN_1 = "ibc/4627AD2524E3E0523047E35BB76CC90E37D9D57ACF14F0FCBCEB2480705F3CB8";
    const MAX_PERCENT = 1;

    let quota: bigint;
    let firstTransfer: bigint;

    const timeout = () => moment.utc().add(5, "minute").unix().toString() + "000000000";

    /** Submits a gov proposal carrying `msg`, votes yes and waits for it to pass. */
    const passProposal = async (msg: any) => {
        let result = await signAndBroadcastTx(val2Wallet, {
            msgs: [new MsgSubmitProposal(
                [msg],
                Coins.fromString("1000000000uluna"),
                val2WalletAddress,
                "metadata",
                "title",
                "summary",
            )],
            chainID: "test-2",
        });
        await blockInclusion();

        let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        expect(txResult.code).toBe(0);
        const proposalId = getValueByIndexAndTypeAndKey(txResult.events, 0, "submit_proposal", "proposal_id");
        expect(proposalId).toBeDefined();

        result = await signAndBroadcastTx(val2Wallet, {
            msgs: [new MsgVote(proposalId, val2WalletAddress, VoteOption.VOTE_OPTION_YES)],
            fee: {
                amount: [{ denom: "uluna", amount: "100000" }],
                gas: "100000",
            },
            chainID: "test-2",
        });
        await blockInclusion();
        await votingPeriod();
        txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        expect(txResult.code).toBe(0);
    };

    const transferToChain1 = (amount: bigint) => signAndBroadcastTx(chain2Wallet, {
        msgs: [new MsgTransfer(
            "transfer",
            CHANNEL,
            new Coin(DENOM, amount.toString()),
            walletAddress,
            walletAddress,
            undefined,
            timeout(),
            undefined,
        )],
        chainID: "test-2",
    });

    test("Must expose the module queries with no rate limit configured", async () => {
        expect(await queryAllRateLimits("test-2")).toStrictEqual([]);
        expect(await queryRateLimit("test-2", CHANNEL, DENOM)).toBeUndefined();
    });

    test("Must add a rate limit through governance", async () => {
        await passProposal(msgAddRateLimit({
            denom: DENOM,
            channelOrClientId: CHANNEL,
            maxPercentSend: MAX_PERCENT.toString(),
            maxPercentRecv: MAX_PERCENT.toString(),
            durationHours: 1,
        }));

        const rateLimit = await pollUntil(
            () => queryRateLimit("test-2", CHANNEL, DENOM),
            (value) => value !== undefined,
        ) as RateLimit;

        expect(rateLimit.path).toStrictEqual({
            denom: DENOM,
            channel_or_client_id: CHANNEL,
        });
        expect(rateLimit.quota).toStrictEqual({
            max_percent_send: MAX_PERCENT.toString(),
            max_percent_recv: MAX_PERCENT.toString(),
            duration_hours: "1",
        });
        expect(rateLimit.flow.inflow).toBe("0");
        expect(rateLimit.flow.outflow).toBe("0");

        // The channel value is the uluna supply snapshotted when the window opened.
        const channelValue = BigInt(rateLimit.flow.channel_value);
        expect(channelValue).toBeGreaterThan(0n);
        quota = (channelValue * BigInt(MAX_PERCENT)) / 100n;
        firstTransfer = quota / 2n;
        expect(firstTransfer).toBeGreaterThan(0n);

        // The suite is only meaningful if the wallet can actually reach the quota.
        const balance = await LCD.chain2.bank.balance(walletAddress);
        expect(BigInt(balance[0].get(DENOM)?.amount.toString() ?? "0")).toBeGreaterThan(quota);

        // Adding a rate limit must not affect the other transfer channels or denoms.
        expect(await queryAllRateLimits("test-2")).toHaveLength(1);
    });

    test("Must let a transfer below the quota through and record the outflow", async () => {
        const result = await transferToChain1(firstTransfer);
        await blockInclusion();

        const txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        expect(txResult.code).toBe(0);

        const rateLimit = await pollUntil(
            () => queryRateLimit("test-2", CHANNEL, DENOM),
            (value) => value?.flow.outflow === firstTransfer.toString(),
        ) as RateLimit;
        expect(rateLimit.flow.outflow).toBe(firstTransfer.toString());
        expect(rateLimit.flow.inflow).toBe("0");
    });

    test("Must reject a transfer that exceeds the quota and leave the flow untouched", async () => {
        // The previous transfer already used half the quota, so a full quota more
        // puts the net outflow at 150% of the threshold.
        const result = await transferToChain1(quota);
        await blockInclusion();

        const txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        expect(txResult.code).not.toBe(0);
        expect(txResult.raw_log).toContain("quota exceeded");

        // The rejected transfer is rolled back, so the flow still only reflects
        // the transfer that succeeded.
        const rateLimit = await queryRateLimit("test-2", CHANNEL, DENOM) as RateLimit;
        expect(rateLimit.flow.outflow).toBe(firstTransfer.toString());
        expect(rateLimit.flow.inflow).toBe("0");
    });

    test("Must record the inflow when the voucher is sent back from chain-1", async () => {
        const amount = firstTransfer / 2n;
        await signAndBroadcastTx(chain1Wallet, {
            msgs: [new MsgTransfer(
                "transfer",
                CHANNEL,
                new Coin(IBC_DENOM_ON_CHAIN_1, amount.toString()),
                walletAddress,
                walletAddress,
                undefined,
                timeout(),
                undefined,
            )],
            chainID: "test-1",
        });
        await ibcTransfer();

        // chain-2 is the source zone here, so the voucher unwinds back to uluna
        // and counts against the same rate limit as an inflow.
        const rateLimit = await pollUntil(
            () => queryRateLimit("test-2", CHANNEL, DENOM),
            (value) => value?.flow.inflow === amount.toString(),
        ) as RateLimit;
        expect(rateLimit.flow.inflow).toBe(amount.toString());
        expect(rateLimit.flow.outflow).toBe(firstTransfer.toString());
    });

    test("Must remove the rate limit through governance", async () => {
        await passProposal(msgRemoveRateLimit({
            denom: DENOM,
            channelOrClientId: CHANNEL,
        }));

        const rateLimit = await pollUntil(
            () => queryRateLimit("test-2", CHANNEL, DENOM),
            (value) => value === undefined,
        );
        expect(rateLimit).toBeUndefined();
        expect(await queryAllRateLimits("test-2")).toStrictEqual([]);

        // With the limit gone the previously rejected transfer must succeed.
        const result = await transferToChain1(quota);
        await blockInclusion();
        const txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        expect(txResult.code).toBe(0);
    });
});
