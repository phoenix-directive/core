package v2_12

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/terra-money/core/v2/app/keepers"
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
		// Included in the start to also run on testnet
		if err := updateValidatorsMinCommissionRate(ctx, k.StakingKeeper); err != nil {
			return nil, err
		}

		if ctx.ChainID() != "phoenix-1" {
			return mm.RunMigrations(ctx, cfg, vm)
		}
		addr := sdk.MustAccAddressFromBech32("")
		multisigAddr := sdk.MustAccAddressFromBech32("")

		// Iterate delegations and unbond all shares
		// burning the coins immediately
		k.StakingKeeper.IterateDelegatorDelegations(ctx, addr, func(d stakingtypes.Delegation) (stop bool) {
			valAddr, err := sdk.ValAddressFromBech32(d.ValidatorAddress)
			if err != nil {
				panic(err)
			}
			// Use this method without adding unbonding to the unbondings queue
			// because it's not necessary to wait for the unbonding period
			// (basically burn the shares and coins immediately)
			_, err = k.StakingKeeper.Unbond(ctx, addr, valAddr, d.Shares)
			if err != nil {
				panic(err)
			}
			return false
		})

		// Given one of the states can be undelegating, we need to iterate over all unbonding delegations
		// and remove them manually to ensure that the undelegated coins are burned.
		bondDenom := k.StakingKeeper.GetParams(ctx).BondDenom
		k.StakingKeeper.IterateDelegatorUnbondingDelegations(ctx, addr, func(ubd stakingtypes.UnbondingDelegation) (stop bool) {
			balances := sdk.NewCoins()
			for i := 0; i < len(ubd.Entries); i++ {
				entry := ubd.Entries[i]
				ubd.RemoveEntry(int64(i))
				i--
				k.StakingKeeper.DeleteUnbondingIndex(ctx, entry.UnbondingId)

				// track undelegation only when remaining or truncated shares are non-zero
				if !entry.Balance.IsZero() {
					amt := sdk.NewCoin(bondDenom, entry.Balance)
					if err := k.BankKeeper.UndelegateCoinsFromModuleToAccount(
						ctx, stakingtypes.NotBondedPoolName, addr, sdk.NewCoins(amt),
					); err != nil {
						panic(err)
					}

					balances = balances.Add(amt)
				}
			}
			k.StakingKeeper.RemoveUnbondingDelegation(ctx, ubd)
			return false
		})

		// Redelegations are two queues but no coins are custodied in any "redelegations_pool",
		// so we can just iterate over all redelegations and remove the indices to prevent issues.
		k.StakingKeeper.IterateDelegatorRedelegations(ctx, addr, func(red stakingtypes.Redelegation) (stop bool) {
			balances := sdk.NewCoins()
			for i := 0; i < len(red.Entries); i++ {
				entry := red.Entries[i]
				red.RemoveEntry(int64(i))
				i--
				k.StakingKeeper.DeleteUnbondingIndex(ctx, entry.UnbondingId)

				if !entry.InitialBalance.IsZero() {
					balances = balances.Add(sdk.NewCoin(bondDenom, entry.InitialBalance))
				}
			}
			k.StakingKeeper.RemoveRedelegation(ctx, red)
			return false
		})

		// Burn all coins in the addr
		k.BankKeeper.IterateAccountBalances(ctx, addr, func(balance sdk.Coin) bool {
			err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, addr, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
			if err != nil {
				panic(err)
			}
			err = k.BankKeeper.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
			if err != nil {
				panic(err)
			}
			return false
		})

		// Burn all coins from the multisig account
		k.BankKeeper.IterateAccountBalances(ctx, multisigAddr, func(balance sdk.Coin) bool {
			err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, multisigAddr, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
			if err != nil {
				panic(err)
			}
			err = k.BankKeeper.BurnCoins(ctx, stakingtypes.NotBondedPoolName, sdk.NewCoins(balance))
			if err != nil {
				panic(err)
			}
			return false
		})

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
	 commission := validator.Commission
	 if commission.MaxRate.LT(sdk.MustNewDecFromStr("0.05")) {
	  commission.MaxRate = sdk.MustNewDecFromStr("0.05")
	  update = true
	 }
	 if commission.Rate.LT(sdk.MustNewDecFromStr("0.05")) {
	  // force update without checking the <24h restriction and the max update rate
	  commission.Rate = sdk.MustNewDecFromStr("0.05")
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
