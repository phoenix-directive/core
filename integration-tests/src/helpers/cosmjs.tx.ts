import { DirectSecp256k1HdWallet, EncodeObject, Registry } from "@cosmjs/proto-signing";
import { SigningStargateClient } from "@cosmjs/stargate";
import { stringToPath } from "@cosmjs/crypto";

const CHAIN_CONFIG = {
    "test-1": {
        rpc: "http://localhost:16657",
        prefix: "terra",
        hdPath: "m/44'/330'/0'/0/0",
    },
    "test-2": {
        rpc: "http://localhost:26657",
        prefix: "terra",
        hdPath: "m/44'/330'/0'/0/0",
    },
} as const;

const DEFAULT_FEE = {
    amount: [{ denom: "uluna", amount: "1500000" }],
    gas: "10000000",
};

type ChainID = keyof typeof CHAIN_CONFIG;
type FeatherMsg = {
    packAny: (isClassic?: boolean) => {
        typeUrl: string;
        value: Uint8Array;
    };
};
type FeatherWallet = {
    key: {
        mnemonic?: string;
        accAddress: (prefix: string) => string;
    };
};

type CachedClient = {
    client: SigningStargateClient;
    registry: Registry;
};

const clientCache = new Map<string, Promise<CachedClient>>();

function isEncodeObject(msg: FeatherMsg | EncodeObject): msg is EncodeObject {
    return "typeUrl" in msg && "value" in msg && !("packAny" in msg);
}

function getWalletMnemonic(wallet: FeatherWallet): string {
    if (typeof wallet.key.mnemonic !== "string") {
        throw new Error("CosmJS transaction helper requires a mnemonic-backed wallet");
    }
    return wallet.key.mnemonic;
}

function registerPrepackedFeatherMsg(registry: Registry, typeUrl: string): void {
    registry.register(typeUrl, {
        create: (value: FeatherMsg) => value,
        encode: (value: FeatherMsg) => ({
            finish: () => value.packAny(false).value,
        }),
        decode: () => {
            throw new Error(`Decoding ${typeUrl} is not supported by the Feather message shim`);
        },
    } as any);
}

async function getClient(wallet: FeatherWallet, chainID: ChainID): Promise<CachedClient> {
    const config = CHAIN_CONFIG[chainID];
    const mnemonic = getWalletMnemonic(wallet);
    const cacheKey = `${chainID}:${mnemonic}`;
    const cached = clientCache.get(cacheKey);
    if (cached) {
        return cached;
    }

    const created = (async () => {
        const registry = new Registry();
        const signer = await DirectSecp256k1HdWallet.fromMnemonic(mnemonic, {
            prefix: config.prefix,
            hdPaths: [stringToPath(config.hdPath)],
        });
        const [account] = await signer.getAccounts();
        expect(account.address).toStrictEqual(wallet.key.accAddress(config.prefix));

        const client = await SigningStargateClient.connectWithSigner(config.rpc, signer, {
            registry,
        });
        return { client, registry };
    })();
    clientCache.set(cacheKey, created);
    return created;
}

const sleep = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

export async function signAndBroadcastTx(
    wallet: FeatherWallet,
    options: {
        msgs: Array<FeatherMsg | EncodeObject>;
        chainID: ChainID;
        memo?: string;
        fee?: typeof DEFAULT_FEE;
    },
): Promise<{ txhash: string }> {
    const cacheKey = `${options.chainID}:${getWalletMnemonic(wallet)}`;
    const { client, registry } = await getClient(wallet, options.chainID);
    const buildMsgs = (targetRegistry: Registry) => options.msgs.map((msg): EncodeObject => {
        if (isEncodeObject(msg)) {
            return msg;
        }

        const any = msg.packAny(false);
        registerPrepackedFeatherMsg(targetRegistry, any.typeUrl);
        return {
            typeUrl: any.typeUrl,
            value: msg,
        };
    });
    let msgs = buildMsgs(registry);

    const signerAddress = wallet.key.accAddress(CHAIN_CONFIG[options.chainID].prefix);
    const fee = options.fee ?? DEFAULT_FEE;
    const memo = options.memo ?? "";
    const broadcast = (signer: SigningStargateClient) => signer.signAndBroadcastSync(
        signerAddress,
        msgs,
        fee,
        memo,
    );

    let txhash: string;
    try {
        txhash = await broadcast(client);
    } catch (e: any) {
        if (!String(e?.message ?? e).includes("account sequence mismatch")) {
            throw e;
        }

        clientCache.delete(cacheKey);
        await sleep(500);
        const fresh = await getClient(wallet, options.chainID);
        msgs = buildMsgs(fresh.registry);
        txhash = await broadcast(fresh.client);
    }

    return { txhash };
}
