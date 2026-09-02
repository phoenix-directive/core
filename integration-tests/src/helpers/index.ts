import {
    SAFE_BLOCK_INCLUSION_TIME,
    SAFE_VOTING_PERIOD_TIME,
    blockInclusion,
    votingPeriod,
    ibcTransfer,
} from "./const"
import { getMnemonics } from "./mnemonics"
import { getLCDClient } from "./lcd.connection"
import { signAndBroadcastTx } from "./cosmjs.tx"

export {
    SAFE_BLOCK_INCLUSION_TIME,
    SAFE_VOTING_PERIOD_TIME,
    blockInclusion,
    votingPeriod,
    ibcTransfer,
    getMnemonics,
    getLCDClient,
    signAndBroadcastTx
}

export const getValueByIndexAndTypeAndKey = (events: any[], index: number, type: string, key: string) => {
    const matchedEvents = events.filter((event: any) =>
        event.type === type &&
        event.attributes.find((attr: any) => attr.key === "msg_index")?.value === index.toString()
    );
    if (matchedEvents.length === 0) {
        throw new Error(`No event found with type ${type} and key ${key}`);
    }
    const matchedAttribute = matchedEvents[0].attributes.find((attr: any) => attr.key === key);
    if (matchedAttribute === undefined) {
        throw new Error(`No attribute found with type ${type} and key ${key}`);
    }
    return matchedAttribute.value;
}

export const getEventsByIndex = (events: any[], index: number) => {
    return events.filter((event: any) => event.attributes.find((attr: any) => attr.key === "msg_index" && attr.value === index.toString()));
}

/**
 * Repeatedly runs `fn` until `predicate` accepts its result. Chain state is
 * eventually consistent from the test's point of view (IBC relaying, callbacks
 * and tx indexing all happen asynchronously), so waiting for a fixed amount of
 * time makes the tests racy. Errors are treated as "not ready yet" because a
 * contract query for a key that does not exist yet answers with a 500.
 *
 * The last attempt is executed outside of the loop so a genuine mismatch still
 * surfaces as a normal assertion diff instead of a timeout.
 */
export const pollUntil = async <T>(
    fn: () => Promise<T>,
    predicate: (value: T) => boolean,
    attempts = 20,
    delay = 500,
): Promise<T> => {
    for (let attempt = 0; attempt < attempts - 1; attempt++) {
        try {
            const value = await fn();
            if (predicate(value)) {
                return value;
            }
        } catch (e) {
            // retry
        }
        await new Promise(resolve => setTimeout(resolve, delay));
    }
    return fn();
}
