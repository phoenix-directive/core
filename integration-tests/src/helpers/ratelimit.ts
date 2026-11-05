/**
 * Helpers for the ratelimit module (github.com/cosmos/ibc-apps/modules/rate-limiting).
 *
 * feather.js does not ship message or query types for this module, so the two
 * governance messages the tests need are encoded by hand and the queries go
 * straight to the LCD REST endpoints exposed by the module's gRPC gateway.
 */

const RATELIMIT_REST_PREFIX = "/Stride-Labs/ibc-rate-limiting/ratelimit";

export const LCD_URL = {
    "test-1": "http://localhost:1316",
    "test-2": "http://localhost:1317",
} as const;

export type ChainID = keyof typeof LCD_URL;

// The gov module account, which is the only authority allowed to change rate limits.
export const GOV_MODULE_ADDRESS = "terra10d07y265gmmuvt4z0w9aw880jnsr700juxf95n";

export type RateLimit = {
    path: { denom: string; channel_or_client_id: string };
    quota: { max_percent_send: string; max_percent_recv: string; duration_hours: string };
    flow: { inflow: string; outflow: string; channel_value: string };
};

// ---------------------------------------------------------------------------
// Minimal protobuf encoding
// ---------------------------------------------------------------------------

function varint(value: number | bigint): number[] {
    let remaining = BigInt(value);
    const out: number[] = [];
    do {
        const byte = Number(remaining & 0x7fn);
        remaining >>= 7n;
        out.push(remaining > 0n ? byte | 0x80 : byte);
    } while (remaining > 0n);
    return out;
}

// Every field used below has a number <= 15, so the tag always fits in one byte.
const tag = (field: number, wireType: number) => (field << 3) | wireType;

function encodeString(field: number, value: string): number[] {
    const bytes = Array.from(Buffer.from(value, "utf8"));
    return [tag(field, 2), ...varint(bytes.length), ...bytes];
}

function encodeUint64(field: number, value: number | bigint): number[] {
    return [tag(field, 0), ...varint(value)];
}

/**
 * feather.js' MsgSubmitProposal only requires its inner messages to expose
 * `packAny()`, so a hand-encoded message can be passed straight through.
 */
const asProposalMessage = (typeUrl: string, bytes: number[]) => ({
    packAny: () => ({ typeUrl, value: new Uint8Array(bytes) }),
});

/**
 * ratelimit.v1.MsgAddRateLimit. `max_percent_send` and `max_percent_recv` are
 * `cosmossdk.io/math.Int` values, which gogoproto encodes as their decimal
 * string representation.
 */
export function msgAddRateLimit(params: {
    denom: string;
    channelOrClientId: string;
    maxPercentSend: string;
    maxPercentRecv: string;
    durationHours: number;
}) {
    return asProposalMessage("/ratelimit.v1.MsgAddRateLimit", [
        ...encodeString(1, GOV_MODULE_ADDRESS),
        ...encodeString(2, params.denom),
        ...encodeString(3, params.channelOrClientId),
        ...encodeString(4, params.maxPercentSend),
        ...encodeString(5, params.maxPercentRecv),
        ...encodeUint64(6, params.durationHours),
    ]);
}

/** ratelimit.v1.MsgRemoveRateLimit */
export function msgRemoveRateLimit(params: { denom: string; channelOrClientId: string }) {
    return asProposalMessage("/ratelimit.v1.MsgRemoveRateLimit", [
        ...encodeString(1, GOV_MODULE_ADDRESS),
        ...encodeString(2, params.denom),
        ...encodeString(3, params.channelOrClientId),
    ]);
}

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

async function get(chainID: ChainID, path: string): Promise<any> {
    const response = await fetch(`${LCD_URL[chainID]}${RATELIMIT_REST_PREFIX}${path}`);
    const body = await response.json();
    if (!response.ok) {
        throw new Error(`${path} answered ${response.status}: ${JSON.stringify(body)}`);
    }
    return body;
}

export async function queryAllRateLimits(chainID: ChainID): Promise<RateLimit[]> {
    const { rate_limits } = await get(chainID, "/ratelimits");
    // The gateway omits the field entirely when no rate limits are configured.
    return rate_limits ?? [];
}

/**
 * Returns the rate limit for a channel/denom pair, or undefined when none is
 * configured. The module answers with an empty `rate_limit` rather than a 404.
 */
export async function queryRateLimit(
    chainID: ChainID,
    channelOrClientId: string,
    denom: string,
): Promise<RateLimit | undefined> {
    const { rate_limit } = await get(
        chainID,
        `/ratelimit/${channelOrClientId}/by_denom?denom=${encodeURIComponent(denom)}`,
    );
    return rate_limit ?? undefined;
}
