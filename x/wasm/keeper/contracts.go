package keeper

import (
	"github.com/terra-money/core/v2/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetExecutedContractAddresses(ctx sdk.Context) (contracts types.ExecutedContracts, found bool) {
	store := k.storeService.OpenKVStore(ctx)
	contractAddressesKey := types.GetExecutedContractsKey()
	b, err := store.Get(contractAddressesKey)
	if err != nil || b == nil {
		return contracts, false
	}

	k.cdc.MustUnmarshal(b, &contracts)
	return contracts, true
}

func (k Keeper) SetExecutedContractAddresses(ctx sdk.Context, contracts types.ExecutedContracts) error {
	store := k.storeService.OpenKVStore(ctx)
	contractAddressesKey := types.GetExecutedContractsKey()
	b := k.cdc.MustMarshal(&contracts)
	err := store.Set(contractAddressesKey, b)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) DeleteExecutedContractAddresses(ctx sdk.Context) error {
	store := k.storeService.OpenKVStore(ctx)
	contractAddressesKey := types.GetExecutedContractsKey()
	err := store.Delete(contractAddressesKey)
	if err != nil {
		return err
	}
	return nil
}
