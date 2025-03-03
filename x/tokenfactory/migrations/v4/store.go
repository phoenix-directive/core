package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/core/v2/x/tokenfactory/types"
)

type Keeper interface {
	GetParams(ctx sdk.Context) types.Params
	SetParams(ctx sdk.Context, params types.Params) error
}

func MigrateStore(ctx sdk.Context, keeper Keeper) error {
	params := keeper.GetParams(ctx)
	params.WhitelistedHooks = []*types.WhitelistedHook{}
	return keeper.SetParams(ctx, params)
}
