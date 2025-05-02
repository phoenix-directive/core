import { getMnemonics, getLCDClient, blockInclusion, getEventsByIndex } from "../../helpers";
import { ContinuousVestingAccount, Coins, MnemonicKey, MsgCreateVestingAccount, Coin, MsgCreatePeriodicVestingAccount, Period } from "@terra-money/feather.js";
import moment from "moment";

describe("Auth Module (https://github.com/terra-money/cosmos-sdk/tree/release/v0.47.x/x/auth)", () => {
    // Prepare environment clients, accounts and wallets
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    const wallet = LCD.chain1.wallet(accounts.genesisVesting1);
    const vestAccAddr1 = accounts.genesisVesting1.accAddress("terra");

    test('Must contain the expected module params', async () => {
        // Query Auth module params
        const moduleParams = await LCD.chain1.auth.parameters("test-1");

        expect(moduleParams)
            .toMatchObject({
                "max_memo_characters": 256,
                "tx_sig_limit": 7,
                "tx_size_cost_per_byte": 10,
                "sig_verify_cost_ed25519": 590,
                "sig_verify_cost_secp256k1": 1000
            });
    });

    test('Must have vesting accounts created on genesis', async () => {
        // Query genesis vesting account info
        const vestAccAddr = accounts.genesisVesting.accAddress("terra");
        const vestAcc = (await LCD.chain1.auth.accountInfo(vestAccAddr)) as ContinuousVestingAccount;

        // Validate the instance of the object
        expect(vestAcc)
            .toBeInstanceOf(ContinuousVestingAccount);
        // Validate the vesting start has been set in the past
        expect(vestAcc.start_time)
            .toBeLessThan(moment().unix());
        // Validate the vesting end has been set in the past
        expect(vestAcc.base_vesting_account.end_time)
            .toBeGreaterThan(moment().unix());
        // Validate the original vesting
        expect(vestAcc.base_vesting_account.original_vesting)
            .toStrictEqual(Coins.fromString("10000000000uluna"));

        // Validate other params from base account
        expect(vestAcc.base_vesting_account.base_account.address)
            .toBe(vestAccAddr);
        expect(vestAcc.getAccountNumber())
            .toBe(4);
        expect(vestAcc.getPublicKey())
            .toBeNull();
        expect(vestAcc.getSequenceNumber())
            .toBe(0);

        // Query the non-vested account balance
        const vestAccBalance = await LCD.chain1.bank.balance(vestAccAddr);

        // Validate the unlocked balance is still available
        expect(vestAccBalance[0].get("uluna"))
            .toStrictEqual(Coin.fromString("1000000000000uluna"));
    });

    test('Must create a random vesting account', async () => {
        const randomAccountAddress = new MnemonicKey().accAddress("terra");
        // Register a new vesting account
        let tx = await wallet.createAndSignTx({
            msgs: [new MsgCreateVestingAccount(
                vestAccAddr1,
                randomAccountAddress,
                Coins.fromString("100uluna"),
                moment().add(1, "minute").unix(),
                false,
            )],
            chainID: "test-1",
        });

        let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        const events = getEventsByIndex(txResult.events, 0);
        expect(events)
            .toEqual([{
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "action",
                    "value": "/cosmos.vesting.v1beta1.MsgCreateVestingAccount"
                }, {
                    "index": true,
                    "key": "sender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "module",
                    "value": "vesting"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "coin_spent",
                "attributes": [{
                    "index": true,
                    "key": "spender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "amount",
                    "value": "100uluna"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "coin_received",
                "attributes": [{
                    "index": true,
                    "key": "receiver",
                    "value": randomAccountAddress
                }, {
                    "index": true,
                    "key": "amount",
                    "value": "100uluna"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "transfer",
                "attributes": [{
                    "index": true,
                    "key": "recipient",
                    "value": randomAccountAddress
                }, {
                    "index": true,
                    "key": "sender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "amount",
                    "value": "100uluna"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "sender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            }])
    });

    test('Must create a periodic vesting account', async () => {
        const randomAccountAddress = new MnemonicKey().accAddress("terra");
        // Register a new vesting account
        let tx = await wallet.createAndSignTx({
            msgs: [new MsgCreatePeriodicVestingAccount (
                vestAccAddr1,
                randomAccountAddress,
                moment().add(1, "minute").unix(),
                [new Period(1000000, "1000000uluna"), new Period(1000000, "1000000uluna")]
            )],
            chainID: "test-1",
        });

        let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        const events = getEventsByIndex(txResult.events, 0);
        expect(events)
            .toEqual([{
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "action",
                    "value": "/cosmos.vesting.v1beta1.MsgCreatePeriodicVestingAccount"
                }, {
                    "index": true,
                    "key": "sender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "module",
                    "value": "vesting"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "coin_spent",
                "attributes": [{
                    "index": true,
                    "key": "spender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "amount",
                    "value": "2000000uluna"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "coin_received",
                "attributes": [{
                    "index": true,
                    "key": "receiver",
                    "value": randomAccountAddress
                }, {
                    "index": true,
                    "key": "amount",
                    "value": "2000000uluna"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "transfer",
                "attributes": [{
                    "index": true,
                    "key": "recipient",
                    "value": randomAccountAddress
                }, {
                    "index": true,
                    "key": "sender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "amount",
                    "value": "2000000uluna"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            },
            {
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "sender",
                    "value": vestAccAddr1
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            }])
    });
});