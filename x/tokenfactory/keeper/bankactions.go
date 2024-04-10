package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	customterratypes "github.com/terra-money/core/v2/x/bank/types"
	"github.com/terra-money/core/v2/x/tokenfactory/types"
)

func (k Keeper) mintTo(ctx sdk.Context, amount sdk.Coin, mintTo string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(mintTo)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
		addr,
		sdk.NewCoins(amount))
}

func (k Keeper) burnFrom(ctx sdk.Context, amount sdk.Coin, burnFrom string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	addr, err := sdk.AccAddressFromBech32(burnFrom)
	if err != nil {
		return err
	}
	coins := sdk.NewCoins(amount)

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		addr,
		types.ModuleName,
		coins)
	if err != nil {
		return err
	}
	recipientAcc := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(customterratypes.ErrUnknownAddress, "module account %s does not exist", recipientAcc))
	}

	err = k.bankKeeper.BlockBeforeSend(ctx, addr, recipientAcc.GetAddress(), coins)
	if err != nil {
		return err
	}
	k.bankKeeper.TrackBeforeSend(ctx, addr, recipientAcc.GetAddress(), coins)

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
}

func (k Keeper) forceTransfer(ctx sdk.Context, amount sdk.Coin, fromAddr string, toAddr string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(amount.Denom)
	if err != nil {
		return err
	}

	fromSdkAddr, err := sdk.AccAddressFromBech32(fromAddr)
	if err != nil {
		return err
	}

	toSdkAddr, err := sdk.AccAddressFromBech32(toAddr)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoins(ctx, fromSdkAddr, toSdkAddr, sdk.NewCoins(amount))
}
