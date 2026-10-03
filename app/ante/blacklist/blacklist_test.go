package blacklist

import (
	"errors"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

type testSignedTx struct {
	authsigning.SigVerifiableTx
	signers [][]byte
	err     error
}

func (tx testSignedTx) GetSigners() ([][]byte, error) {
	return tx.signers, tx.err
}

type testUnsignedTx struct {
	sdk.Tx
}

func TestBlacklistDecorator(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("terra", "terrapub")
	if !Blacklist["terra14m8unq627j97ys8f8k0vlm6nukuwx5wd0zt7q5m"] {
		t.Fatal("specified victim address is missing from the blacklist")
	}
	addresses := map[string]string{
		"victim":   "terra1kvwkvurw4xexw69ef772p95jajnjjjsq6d5uca",
		"attacker": "terra1agp4wwzgn6fuxqrsqgjvfhvmfqf63ekvn9nvyh",
	}
	for role, address := range addresses {
		if !Blacklist[address] {
			t.Fatalf("%s address %s is missing from the blacklist", role, address)
		}
	}

	allowed := sdk.AccAddress(make([]byte, 20))
	decorator := NewBlacklistDecorator()
	ctx := sdk.Context{}
	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	for _, address := range addresses {
		blocked, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			t.Fatal(err)
		}
		called = false
		_, err = decorator.AnteHandle(ctx, testSignedTx{signers: [][]byte{allowed, blocked}}, false, next)
		if err == nil || !strings.Contains(err.Error(), "is blacklisted") || called {
			t.Fatalf("blacklisted signer %s was not rejected: err=%v, next called=%t", address, err, called)
		}
	}

	_, err := decorator.AnteHandle(ctx, testSignedTx{signers: [][]byte{allowed}}, false, next)
	if err != nil || !called {
		t.Fatalf("allowed signer was not passed to next decorator: err=%v, next called=%t", err, called)
	}

	called = false
	signerErr := errors.New("invalid signers")
	_, err = decorator.AnteHandle(ctx, testSignedTx{err: signerErr}, false, next)
	if !errors.Is(err, signerErr) || called {
		t.Fatalf("signer error was not returned: err=%v, next called=%t", err, called)
	}

	_, err = decorator.AnteHandle(ctx, testUnsignedTx{}, false, next)
	if err != nil || !called {
		t.Fatalf("unsigned transaction was not passed to next decorator: err=%v, next called=%t", err, called)
	}
}
