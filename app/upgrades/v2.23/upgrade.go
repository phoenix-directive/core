package v2_23

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/terra-money/core/v2/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k keepers.TerraAppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// The ratelimit module is added in this upgrade. It starts with no rate limits
		// configured, so IBC transfers are unaffected until limits are added by governance
		// (MsgAddRateLimit).
		return mm.RunMigrations(ctx, cfg, vm)
	}
}
