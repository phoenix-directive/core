import { getMnemonics } from "../../helpers/mnemonics";
import { getLCDClient } from "../../helpers/lcd.connection";
import { StakeAuthorization, MsgGrantAuthorization, AuthorizationGrant, Coin, MsgExecAuthorized, MsgDelegate } from "@terra-money/feather.js";
import { AuthorizationType } from "@terra-money/terra.proto/cosmos/staking/v1beta1/authz";
import moment from "moment";
import { blockInclusion } from "../../helpers/const";
import { getEventsByIndex, signAndBroadcastTx } from "../../helpers";

describe("Authz Module", () => {
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    // Accounts used in chain2, which means that 
    // will not cause conflicts with txs nonces
    const granterWallet = LCD.chain2.wallet(accounts.feeshareMnemonic);
    const granteeWallet = LCD.chain2.wallet(accounts.pobMnemonic);
    const granterAddr = accounts.feeshareMnemonic.accAddress("terra");
    const granteeAddr = accounts.pobMnemonic.accAddress("terra");
    const val2Addr = accounts.val2.valAddress("terra");

    test('Must register the granter', async () => {
        let result = await signAndBroadcastTx(granterWallet, {
            msgs: [new MsgGrantAuthorization(
                granterAddr,
                granteeAddr,
                new AuthorizationGrant(
                    new StakeAuthorization(
                        AuthorizationType.AUTHORIZATION_TYPE_DELEGATE,
                        Coin.fromString("1000000uluna"),
                    ),
                    moment().add(1, "hour").toDate(),
                ),
            )],
            chainID: "test-2",
        });
        await blockInclusion();

        // Check the MsgGrantAuthorization executed as expected 
        let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
        const events = getEventsByIndex(txResult.events, 0);
        expect(events)
            .toStrictEqual([{
                "type": "message",
                "attributes": [{
                    "index": true,
                    "key": "action",
                    "value": "/cosmos.authz.v1beta1.MsgGrant"
                }, {
                    "index": true,
                    "key": "sender",
                    "value": "terra120rzk7n6cd2vufkmwrat34adqh0rgca9tkyfe5"
                }, {
                    "index": true,
                    "key": "module",
                    "value": "authz"
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            }, {
                "type": "cosmos.authz.v1beta1.EventGrant",
                "attributes": [{
                    "index": true,
                    "key": "grantee",
                    "value": "\"terra1v0eee20gjl68fuk0chyrkch2z7suw2mhg3wkxf\""
                }, {
                    "index": true,
                    "key": "granter",
                    "value": "\"terra120rzk7n6cd2vufkmwrat34adqh0rgca9tkyfe5\""
                }, {
                    "index": true,
                    "key": "msg_type_url",
                    "value": "\"/cosmos.staking.v1beta1.MsgDelegate\""
                }, {
                    "index": true,
                    "key": "msg_index",
                    "value": "0"
                }]
            }]);
    });

    describe("Grantee must execute", () => {
        test("delegation on belhalf of granter", async () => {
            let result = await signAndBroadcastTx(granteeWallet, {
                msgs: [new MsgExecAuthorized(
                    granteeAddr,
                    [new MsgDelegate(
                        granterAddr,
                        val2Addr,
                        Coin.fromString("1000000uluna"),
                    )]
                )],
                chainID: "test-2",
            });
            await blockInclusion();

            let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
            const events = getEventsByIndex(txResult.events, 0);
            let latestIndex = events.length - 1;

            expect(events[0])
                .toStrictEqual({
                    "type": "message",
                    "attributes": [{
                        "index": true,
                        "key": "action",
                        "value": "/cosmos.authz.v1beta1.MsgExec"
                    }, {
                        "index": true,
                        "key": "sender",
                        "value": "terra1v0eee20gjl68fuk0chyrkch2z7suw2mhg3wkxf"
                    }, {
                        "index": true,
                        "key": "module",
                        "value": "authz"
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
            expect(events[1])
                .toStrictEqual({
                    "type": "cosmos.authz.v1beta1.EventRevoke",
                    "attributes": [{
                        "index": true,
                        "key": "grantee",
                        "value": "\"terra1v0eee20gjl68fuk0chyrkch2z7suw2mhg3wkxf\""
                    }, {
                        "index": true,
                        "key": "granter",
                        "value": "\"terra120rzk7n6cd2vufkmwrat34adqh0rgca9tkyfe5\""
                    }, {
                        "index": true,
                        "key": "msg_type_url",
                        "value": "\"/cosmos.staking.v1beta1.MsgDelegate\""
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });

            expect(events[latestIndex])
                .toStrictEqual({
                    "type": "delegate",
                    "attributes": [{
                        "index": true,
                        "key": "validator",
                        "value": "terravaloper1llgzglr9yyy4gyjh8p5kepgm5wyl358de47rqk"
                    }, {
                        "index": true,
                        "key": "delegator",
                        "value": "terra120rzk7n6cd2vufkmwrat34adqh0rgca9tkyfe5"
                    }, {
                        "index": true,
                        "key": "amount",
                        "value": "1000000uluna"
                    }, {
                        "index": true,
                        "key": "new_shares",
                        "value": "1000000.000000000000000000"
                    }, {
                        "index": true,
                        "key": "authz_msg_index",
                        "value": "0"
                    }, {
                        "index": true,
                        "key": "msg_index",
                        "value": "0"
                    }]
                });
        });
    })
});
