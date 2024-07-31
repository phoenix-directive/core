package blacklist

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var Blacklist = map[string]bool{
	// IBC exploiter
	"terra1wrve5z5vsmrgy6ldcveq93aldr6wk3qmxavs4j": true,
}

type BlacklistAnteHandler struct {
}

func NewBlacklistDecorator() BlacklistAnteHandler {
	return BlacklistAnteHandler{}
}

func (b BlacklistAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if ok {
		for _, sig := range sigTx.GetSigners() {
			if Blacklist[sig.String()] {
				return ctx, fmt.Errorf("signer %s is blacklisted", sig.String())
			}
		}
	}
	return next(ctx, tx, simulate)
}
