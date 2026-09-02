import { LCDClient } from "@terra-money/feather.js";

// Transactions are broadcasted in "sync" mode, which returns as soon as the
// node accepts them into the mempool. The tests then sleep for a single block
// (see blockInclusion) before reading the transaction back. When a transaction
// lands one block later than expected the LCD answers with a 404 and the test
// fails with an opaque AxiosError, so poll for the transaction instead of
// failing on the first miss.
const TX_INFO_MAX_ATTEMPTS = 20;
const TX_INFO_RETRY_DELAY = 500;

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

const isNotFound = (e: any) => e?.response?.status === 404 || e?.code === "ECONNREFUSED";

function withTxInfoRetry(client: LCDClient): LCDClient {
    const txInfo = client.tx.txInfo.bind(client.tx);

    client.tx.txInfo = async (...args: Parameters<typeof txInfo>) => {
        let lastError: any;
        for (let attempt = 0; attempt < TX_INFO_MAX_ATTEMPTS; attempt++) {
            try {
                return await txInfo(...args);
            } catch (e: any) {
                if (!isNotFound(e)) {
                    throw e;
                }
                lastError = e;
                await sleep(TX_INFO_RETRY_DELAY);
            }
        }
        throw new Error(
            `Transaction ${args[0]} was not found after ${TX_INFO_MAX_ATTEMPTS * TX_INFO_RETRY_DELAY}ms: ${lastError?.message}`,
        );
    };

    return client;
}

export function getLCDClient() {
    return {
        chain1: withTxInfoRetry(new LCDClient({
            "test-1": {
                lcd: "http://localhost:1316",
                chainID: "test-1",
                gasPrices: "0.15uluna",
                gasAdjustment: 1.5,
                prefix: "terra"
            }
        })),
        chain2: withTxInfoRetry(new LCDClient({
            "test-2": {
                lcd: "http://localhost:1317",
                chainID: "test-2",
                gasPrices: "0.15uluna",
                gasAdjustment: 1.5,
                prefix: "terra"
            }
        }))
    }
}
