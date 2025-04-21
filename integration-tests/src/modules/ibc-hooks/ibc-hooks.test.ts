import { Coin, Coins, MsgInstantiateContract, MsgStoreCode, MsgTransfer } from "@terra-money/feather.js";
import { deriveIbcHooksSender } from "@terra-money/feather.js/dist/core/ibc-hooks";
import { ibcTransfer, getMnemonics, getLCDClient, blockInclusion, getValueByIndexAndTypeAndKey, getEventsByIndex } from "../../helpers";
import fs from "fs";
import path from 'path';
import moment from "moment";
// import { Height } from "@terra-money/feather.js/dist/core/ibc/core/client/Height";

describe("IbcHooks Module (github.com/cosmos/ibc-apps/modules/ibc-hooks/v7) ", () => {
    // Prepare the LCD and wallets. chain1Wallet is the one that will
    // deploy the contract on chain 1 and chain2Wallet will be used 
    // to send IBC messages from chain 2 to interact with the contract.
    const LCD = getLCDClient();
    const accounts = getMnemonics();
    const chain1Wallet = LCD.chain1.wallet(accounts.ibcHooksMnemonic);
    const chain2Wallet = LCD.chain2.wallet(accounts.ibcHooksMnemonic);
    const walletAddress = accounts.ibcHooksMnemonic.accAddress("terra");
    const derivedHooksWalletAddress = deriveIbcHooksSender("channel-0", walletAddress, "terra");
    let contractAddress: string;

    // Read the counter contract, store on chain, 
    // instantiate to be used in the following tests
    // and finally save the contract address.
    beforeAll(async () => {
        let tx = await chain1Wallet.createAndSignTx({
            msgs: [new MsgStoreCode(
                walletAddress,
                fs.readFileSync(path.join(__dirname, "/../../contracts/counter.wasm")).toString("base64"),
            )],
            chainID: "test-1",
        });

        let result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        let txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        let codeId = getValueByIndexAndTypeAndKey(txResult.events, 0, "store_code", "code_id");
        expect(codeId).toBeDefined();

        const msgInstantiateContract = new MsgInstantiateContract(
            walletAddress,
            walletAddress,
            codeId,
            { count: 0 },
            Coins.fromString("1uluna"),
            "counter contract " + Math.random(),
        );

        tx = await chain1Wallet.createAndSignTx({
            msgs: [msgInstantiateContract],
            chainID: "test-1",
        });
        result = await LCD.chain1.tx.broadcastSync(tx, "test-1");
        await blockInclusion();
        txResult = await LCD.chain1.tx.txInfo(result.txhash, "test-1") as any;
        contractAddress = getValueByIndexAndTypeAndKey(txResult.events, 0, "instantiate", "_contract_address");
        expect(contractAddress).toBeDefined();
    })

    // This test send an IBC message to **chain-2** which is **relayed** to **chain-1** 
    // The flow represents a successful MsgTransfe with IBCHooks request to the smart contrat.
    describe("Should execute hooks from chain 2 to chain 1", () => {
        test('must increment the counter successfully', async () => {
            let tx = await chain2Wallet.createAndSignTx({
                msgs: [
                    new MsgTransfer(
                    "transfer",
                    "channel-0",
                    Coin.fromString("1uluna"),
                    walletAddress,
                    contractAddress,
                    undefined,
                    moment.utc().add(1, "minute").unix().toString() + "000000000",
                    `{"wasm":{"contract": "${contractAddress}" ,"msg": {"increment": {}}}}`
                ),
                new MsgTransfer(
                    "transfer",
                    "channel-0",
                    Coin.fromString("1uluna"),
                    walletAddress,
                    contractAddress,
                    undefined,
                    moment.utc().add(1, "minute").unix().toString() + "000000000",
                    `{"wasm":{"contract": "${contractAddress}" ,"msg": {"increment": {}}}}`
                ),
                ],
                chainID: "test-2",
            });
            let result = await LCD.chain2.tx.broadcastSync(tx, "test-2");
            await ibcTransfer();
            let txResult = await LCD.chain2.tx.txInfo(result.txhash, "test-2") as any;
            let txEvents = getEventsByIndex(txResult.events, 0);
            expect(txEvents.find((event: any) => event.type === "ibc_transfer"))
                .toStrictEqual({
                    type: "ibc_transfer",
                    attributes: [
                        {index: true, key: "sender", value: walletAddress},
                        {index: true, key: "receiver", value: contractAddress},
                        {index: true, key: "amount", value: "1"},
                        {index: true, key: "denom", value: "uluna"},
                        {index: true, key: "memo", value: `{"wasm":{"contract": "${contractAddress}" ,"msg": {"increment": {}}}}`},
                        {index: true, key: "msg_index", value: "0"}
                    ]
                })
            // query to validate the count is 1
            let res = await LCD.chain1.wasm.contractQuery(
                contractAddress,
                { "get_count": { "addr": derivedHooksWalletAddress } }
            );
            expect(res).toStrictEqual({ "count": 1 });

            // query to validate the count is 1
            res = await LCD.chain1.wasm.contractQuery(
                contractAddress,
                { "get_total_funds": { "addr": derivedHooksWalletAddress } }
            );
            expect(res).toStrictEqual({
                "total_funds": [{
                    "denom": "ibc/4627AD2524E3E0523047E35BB76CC90E37D9D57ACF14F0FCBCEB2480705F3CB8",
                    "amount": "2"
                }]
            });
        });
    })

    // This test send an IBC message to **chain-1** which is **relayed** to **chain-2** 
    // with an acknowledgement callback for chain-1.
    // The flow represents a successful MsgTransfer with callback to the smart contract.
    describe("Must execute hooks callback from chain 1 to chain 2", () => {
        test('increment the counter on successful callback', async () => {
                let tx = await chain1Wallet.createAndSignTx({
                    msgs: [
                        new MsgTransfer(
                            "transfer",
                            "channel-0",
                            Coin.fromString("1uluna"),
                            walletAddress,
                            derivedHooksWalletAddress,
                            undefined,
                            moment.utc().add(10, "second").unix().toString() + "000000000",
                            `{"ibc_callback": "${contractAddress}"}`
                        ),
                        new MsgTransfer(
                            "transfer",
                            "channel-0",
                            Coin.fromString("1uluna"),
                            walletAddress,
                            derivedHooksWalletAddress,
                            undefined,
                            moment.utc().add(10, "second").unix().toString() + "000000000",
                            `{"ibc_callback": "${contractAddress}"}`
                        ),
                    ],
                    chainID: "test-1",
                });
                await LCD.chain1.tx.broadcastSync(tx, "test-1")
                await ibcTransfer();
                await blockInclusion();
                await blockInclusion();
                let res = await LCD.chain1.wasm.contractQuery(
                    contractAddress,
                    { "get_count": { "addr": contractAddress } }
                );
                expect(res).toStrictEqual({ "count": 2 });
                res = await LCD.chain1.wasm.contractQuery(
                    contractAddress,
                    { "get_total_funds": { "addr": contractAddress } }
                );
                expect(res).toStrictEqual({ "total_funds": [] });
        });
    })


    // This test send an IBC message to **chain-1** which is **NOT relayed** because of timeout.
    // The flow represents a failed MsgTransfer with callback to the smart contract.
    describe("Must execute hooks callback on chain 1", () => {
        test('with a timeout of -1 second', async () => {
            let tx = await chain1Wallet.createAndSignTx({
                msgs: [
                    new MsgTransfer(
                        "transfer",
                        "channel-0",
                        Coin.fromString("1uluna"),
                        walletAddress,
                        derivedHooksWalletAddress,
                        undefined,
                        moment.utc().add(-1, "second").unix().toString() + "000000000",
                        `{"ibc_callback": "${contractAddress}"}`
                    ),
                    new MsgTransfer(
                        "transfer",
                        "channel-0",
                        Coin.fromString("1uluna"),
                        walletAddress,
                        derivedHooksWalletAddress,
                        undefined,
                        moment.utc().add(-1, "second").unix().toString() + "000000000",
                        `{"ibc_callback": "${contractAddress}"}`
                    ),
                ],
                chainID: "test-1",
            });
            await LCD.chain1.tx.broadcastSync(tx, "test-1")
            await ibcTransfer();
            await ibcTransfer();
            let res = await LCD.chain1.wasm.contractQuery(
                contractAddress,
                { "get_count": { "addr": contractAddress } }
            );
            expect(res).toStrictEqual({ "count": 22 });
            res = await LCD.chain1.wasm.contractQuery(
                contractAddress,
                { "get_total_funds": { "addr": contractAddress } }
            );
            expect(res).toStrictEqual({ "total_funds": [] });
        })
    });
});