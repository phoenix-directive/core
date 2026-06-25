package keepers

import (

	// #nosec G702

	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"

	ibchookstypes "github.com/cosmos/ibc-apps/modules/ibc-hooks/v10/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	tokenfactorytypes "github.com/terra-money/core/v2/x/tokenfactory/types"

	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
	feesharetypes "github.com/terra-money/core/v2/x/feeshare/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/terra-money/core/v2/client/docs/statik"
)

func (keepers *TerraAppKeepers) GenerateKeys() {
	keepers.keys = storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibcexported.StoreKey,
		upgradetypes.StoreKey, evidencetypes.StoreKey, ibctransfertypes.StoreKey,
		authzkeeper.StoreKey, feegrant.StoreKey,
		icahosttypes.StoreKey, icacontrollertypes.StoreKey, packetforwardtypes.StoreKey,
		consensusparamtypes.StoreKey, tokenfactorytypes.StoreKey, wasmtypes.StoreKey,
		ibchookstypes.StoreKey, crisistypes.StoreKey,
		alliancetypes.StoreKey, feesharetypes.StoreKey,
	)

	keepers.tkeys = storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
}

func (keepers *TerraAppKeepers) GetKVStoreKey() map[string]*storetypes.KVStoreKey {
	return keepers.keys
}

func (keepers *TerraAppKeepers) GetTransientStoreKey() map[string]*storetypes.TransientStoreKey {
	return keepers.tkeys
}

func (keepers *TerraAppKeepers) GetMemoryStoreKey() map[string]*storetypes.MemoryStoreKey {
	return keepers.memKeys
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (keepers *TerraAppKeepers) GetKey(storeKey string) *storetypes.KVStoreKey {
	return keepers.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (keepers *TerraAppKeepers) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return keepers.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (keepers *TerraAppKeepers) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return keepers.memKeys[storeKey]
}
