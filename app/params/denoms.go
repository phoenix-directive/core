package params

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/terra-money/core/v2/app/config"
)

func RegisterDenomsConfig() error {
	err := sdk.RegisterDenom(config.MicroLuna, math.LegacyNewDecWithPrec(1, 6))
	if err != nil {
		return err
	}
	return nil
}
