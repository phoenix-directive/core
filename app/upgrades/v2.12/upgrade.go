package v2_12

import (
	"fmt"

	"cosmossdk.io/math"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/terra-money/core/v2/app/keepers"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distributionkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

type EscrowUpdate struct {
	EscrowAddress sdk.AccAddress
	Assets        []sdk.Coin
}

// To test this upgrade handler set the following address "terra1v0eee20gjl68fuk0chyrkch2z7suw2mhg3wkxf"
// on the variables below: addr and multisigAddr.
// then run: npm run test:chain:upgrade:v12
func CreateUpgradeHandler(
	mm *module.Manager,
	cfg module.Configurator,
	k keepers.TerraAppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		if err := updateValidatorsMinCommissionRate(ctx, k.StakingKeeper); err != nil {
			return nil, err
		}

		var addr sdk.AccAddress
		var multisigAddr sdk.AccAddress

		if ctx.ChainID() != "phoenix-1" {
			addr = sdk.MustAccAddressFromBech32("")
			multisigAddr = sdk.MustAccAddressFromBech32("")
		} else {
			addr = sdk.MustAccAddressFromBech32("terra1wd8tc98um0x6c9l46vhg00gudzgleefl6tvshd")
			multisigAddr = sdk.MustAccAddressFromBech32("terra1xduqpf6aah0nftppuez7upl6curmykl3cxdek4h5wacw7hn0fr9sr029ze")
		}

		if err := burnTokensFromAccount(ctx, k.StakingKeeper, k.BankKeeper, k.DistrKeeper, k.AccountKeeper, addr); err != nil {
			return nil, err
		}
		if err := burnTokensFromAccount(ctx, k.StakingKeeper, k.BankKeeper, k.DistrKeeper, k.AccountKeeper, multisigAddr); err != nil {
			return nil, err
		}

		return mm.RunMigrations(ctx, cfg, vm)
	}
}

func updateValidatorsMinCommissionRate(ctx sdk.Context, sk *stakingkeeper.Keeper) error {
	// Update min commission rate for new / validators who are updating
	stakingParams := sk.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.MustNewDecFromStr("0.05")
	if err := sk.SetParams(ctx, stakingParams); err != nil {
		return err
	}

	// Update all validators to have a min commission rate of 5%
	validators := sk.GetAllValidators(ctx)
	for _, validator := range validators {
		update := false
		if validator.Commission.MaxRate.LT(sdk.MustNewDecFromStr("0.05")) {
			validator.Commission.MaxRate = sdk.MustNewDecFromStr("0.05")
			update = true
		}
		if validator.Commission.Rate.LT(sdk.MustNewDecFromStr("0.05")) {
			// force update without checking the <24h restriction and the max update rate
			validator.Commission.Rate = sdk.MustNewDecFromStr("0.05")
			update = true
		}
		if update {
			validator.Commission.UpdateTime = ctx.BlockTime()
			if err := sk.Hooks().BeforeValidatorModified(ctx, validator.GetOperator()); err != nil {
				return err
			}
			sk.SetValidator(ctx, validator)
		}
	}
	return nil
}

func burnTokensFromAccount(ctx sdk.Context, sk *stakingkeeper.Keeper, bk bankkeeper.Keeper, dk distributionkeeper.Keeper, ak accountkeeper.AccountKeeper, addr sdk.AccAddress) error {
	acc := ak.GetAccount(ctx, addr)
	if acc == nil {
		return fmt.Errorf("account %s not found", addr)
	}
	// Iterate delegations and unbond all shares
	// burning the coins immediately
	bondDenom := sk.GetParams(ctx).BondDenom
	var err error
	sk.IterateDelegatorDelegations(ctx, addr, func(d stakingtypes.Delegation) (stop bool) {
		var valAddr sdk.ValAddress
		valAddr, err = sdk.ValAddressFromBech32(d.ValidatorAddress)
		if err != nil {
			return true
		}

		// Withdraw delegation rewards first
		_, err = dk.WithdrawDelegationRewards(ctx, addr, valAddr)
		if err != nil {
			return true
		}
		// Use this method without adding unbonding to the unbondings queue
		// because it's not necessary to wait for the unbonding period
		var unbondedAmount math.Int
		unbondedAmount, err = sk.Unbond(ctx, addr, valAddr, d.Shares)
		if err != nil {
			return true
		}

		// After unbonding, burn the coins depending on the validator's status
		validator := sk.Validator(ctx, valAddr)
		if validator.IsBonded() {
			if err = bk.BurnCoins(ctx, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(bondDenom, unbondedAmount))); err != nil {
				return true
			}
		} else {
			if err = bk.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(sdk.NewCoin(bondDenom, unbondedAmount))); err != nil {
				return true
			}
		}

		return false
	})
	if err != nil {
		return err
	}

	// Given one of the states can be undelegating, we need to iterate over all unbonding delegations
	// and remove them manually to ensure that the undelegated coins are burned.
	sk.IterateDelegatorUnbondingDelegations(ctx, addr, func(ubd stakingtypes.UnbondingDelegation) (stop bool) {
		for i := 0; i < len(ubd.Entries); i++ {
			entry := ubd.Entries[i]
			ubd.RemoveEntry(int64(i))
			i--
			sk.DeleteUnbondingIndex(ctx, entry.UnbondingId)

			// track undelegation only when remaining or truncated shares are non-zero
			if !entry.Balance.IsZero() {
				amt := sdk.NewCoin(bondDenom, entry.Balance)
				if err = bk.BurnCoins(
					ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(amt),
				); err != nil {
					return true
				}
			}
		}
		sk.RemoveUnbondingDelegation(ctx, ubd)
		return false
	})
	if err != nil {
		return err
	}

	// Redelegations are two queues but no coins are custodied in any "redelegations_pool",
	// so we can just iterate over all redelegations and remove the indices to prevent issues.
	sk.IterateDelegatorRedelegations(ctx, addr, func(red stakingtypes.Redelegation) (stop bool) {
		for i := 0; i < len(red.Entries); i++ {
			entry := red.Entries[i]
			red.RemoveEntry(int64(i))
			i--
			sk.DeleteUnbondingIndex(ctx, entry.UnbondingId)
		}
		sk.RemoveRedelegation(ctx, red)
		return false
	})

	// Set account back to a base account before burning to vest everything
	switch vestingAcc := acc.(type) {
	case *types.ContinuousVestingAccount:
		ak.SetAccount(ctx, vestingAcc.BaseVestingAccount)
	case *types.DelayedVestingAccount:
		ak.SetAccount(ctx, vestingAcc.BaseVestingAccount)
	case *types.PeriodicVestingAccount:
		ak.SetAccount(ctx, vestingAcc.BaseVestingAccount)
	default:
		// do nothing
	}

	// Burn all coins in the addr
	bk.IterateAccountBalances(ctx, addr, func(balance sdk.Coin) bool {
		err = bk.SendCoinsFromAccountToModule(ctx, addr, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
		if err != nil {
			return true
		}
		err = bk.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
		if err != nil {
			return true
		}
		return false
	})
	return err
}
