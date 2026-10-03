package blacklist

import (
	"errors"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/terra-money/core/v2/app/params"
)

type testSignedTx struct {
	authsigning.SigVerifiableTx
	signers [][]byte
	msgs    []sdk.Msg
	err     error
}

func (tx testSignedTx) GetSigners() ([][]byte, error) {
	return tx.signers, tx.err
}

func (tx testSignedTx) GetMsgs() []sdk.Msg {
	return tx.msgs
}

type testUnsignedTx struct {
	sdk.Tx
}

func (testUnsignedTx) GetMsgs() []sdk.Msg {
	return nil
}

type testFeeTx struct {
	sdk.FeeTx
	granter []byte
}

func (tx testFeeTx) FeeGranter() []byte {
	return tx.granter
}

func (testFeeTx) GetMsgs() []sdk.Msg {
	return nil
}

func TestBlacklistDecorator(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("terra", "terrapub")
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
	decorator := NewBlacklistDecorator(params.MakeEncodingConfig().Marshaler)
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

func TestBlacklistDecoratorAuthz(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("terra", "terrapub")
	blocked, err := sdk.AccAddressFromBech32("terra1kvwkvurw4xexw69ef772p95jajnjjjsq6d5uca")
	if err != nil {
		t.Fatal(err)
	}
	grantee := sdk.AccAddress(make([]byte, 20))
	decorator := NewBlacklistDecorator(params.MakeEncodingConfig().Marshaler)
	ctx := sdk.Context{}
	nextCalled := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		nextCalled = true
		return ctx, nil
	}

	for _, depth := range []int{0, 1, 11} {
		inner := &banktypes.MsgSend{FromAddress: blocked.String(), ToAddress: grantee.String()}
		exec := authz.NewMsgExec(grantee, []sdk.Msg{inner})
		for i := 0; i < depth; i++ {
			innerExec := exec
			exec = authz.NewMsgExec(grantee, []sdk.Msg{&innerExec})
		}
		nextCalled = false
		_, err = decorator.AnteHandle(ctx, testSignedTx{
			signers: [][]byte{grantee},
			msgs:    []sdk.Msg{&exec},
		}, false, next)
		if err == nil || !strings.Contains(err.Error(), "is blacklisted") || nextCalled {
			t.Fatalf("authz bypass was not rejected (depth=%d): err=%v, next called=%t", depth, err, nextCalled)
		}
	}

	allowed := &banktypes.MsgSend{FromAddress: grantee.String(), ToAddress: grantee.String()}
	exec := authz.NewMsgExec(grantee, []sdk.Msg{allowed})
	for i := 0; i < 11; i++ {
		innerExec := exec
		exec = authz.NewMsgExec(grantee, []sdk.Msg{&innerExec})
	}
	nextCalled = false
	_, err = decorator.AnteHandle(ctx, testSignedTx{
		signers: [][]byte{grantee},
		msgs:    []sdk.Msg{&exec},
	}, false, next)
	if err != nil || !nextCalled {
		t.Fatalf("allowed deeply nested authz message was rejected: err=%v, next called=%t", err, nextCalled)
	}
}

func TestBlacklistDecoratorFeeGranter(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("terra", "terrapub")
	blocked, err := sdk.AccAddressFromBech32("terra1kvwkvurw4xexw69ef772p95jajnjjjsq6d5uca")
	if err != nil {
		t.Fatal(err)
	}
	allowed := sdk.AccAddress(make([]byte, 20))
	decorator := NewBlacklistDecorator(params.MakeEncodingConfig().Marshaler)
	called := false
	next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		called = true
		return ctx, nil
	}

	for _, tc := range []struct {
		name    string
		granter []byte
		blocked bool
	}{
		{name: "blacklisted", granter: blocked, blocked: true},
		{name: "allowed", granter: allowed},
		{name: "absent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			called = false
			_, err := decorator.AnteHandle(sdk.Context{}, testFeeTx{granter: tc.granter}, false, next)
			if tc.blocked {
				if err == nil || !strings.Contains(err.Error(), "fee granter "+blocked.String()+" is blacklisted") || called {
					t.Fatalf("blacklisted fee granter was not rejected: err=%v, next called=%t", err, called)
				}
			} else if err != nil || !called {
				t.Fatalf("allowed fee granter was rejected: err=%v, next called=%t", err, called)
			}
		})
	}
}

func TestBlacklistDecoratorTestChainAddress(t *testing.T) {
	sdk.GetConfig().SetBech32PrefixForAccount("terra", "terrapub")
	address, err := sdk.AccAddressFromBech32("terra1a698u5rm2x6y50x5m3q37tnn0k6d4rjpfc8e7h")
	if err != nil {
		t.Fatal(err)
	}
	decorator := NewBlacklistDecorator(params.MakeEncodingConfig().Marshaler)

	for _, tc := range []struct {
		chainID string
		blocked bool
	}{
		{chainID: "blacklist-v222-test-1", blocked: true},
		{chainID: "phoenix-1", blocked: false},
	} {
		t.Run(tc.chainID, func(t *testing.T) {
			called := false
			next := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
				called = true
				return ctx, nil
			}
			_, err := decorator.AnteHandle(sdk.Context{}.WithChainID(tc.chainID), testSignedTx{
				signers: [][]byte{address},
			}, false, next)
			if tc.blocked && (err == nil || !strings.Contains(err.Error(), "is blacklisted") || called) {
				t.Fatalf("test address was not blocked on %s: err=%v, next called=%t", tc.chainID, err, called)
			}
			if !tc.blocked && (err != nil || !called) {
				t.Fatalf("test address was blocked on %s: err=%v, next called=%t", tc.chainID, err, called)
			}
		})
	}
}
