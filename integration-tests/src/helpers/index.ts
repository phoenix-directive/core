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
