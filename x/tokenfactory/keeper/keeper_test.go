package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/terra-money/core/v2/app/test_helpers"
	"github.com/terra-money/core/v2/x/tokenfactory/keeper"
	"github.com/terra-money/core/v2/x/tokenfactory/types"
)

type KeeperTestSuite struct {
	test_helpers.AppTestSuite

	queryClient    types.QueryClient
	msgServer      types.MsgServer
	contractKeeper wasmtypes.ContractOpsKeeper
	bankMsgServer  banktypes.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	s.Setup()
	s.contractKeeper = wasmkeeper.NewGovPermissionKeeper(s.App.Keepers.WasmKeeper)
	s.queryClient = types.NewQueryClient(s.QueryHelper)
	s.msgServer = keeper.NewMsgServerImpl(s.App.Keepers.TokenFactoryKeeper)
	s.bankMsgServer = bankkeeper.NewMsgServerImpl(s.App.Keepers.BankKeeper)
}

func (s *KeeperTestSuite) TestCreateModuleAccount() {
	// setup new next account number
	nextAccountNumber := s.App.Keepers.AccountKeeper.NextAccountNumber(s.Ctx)

	// ensure module account was removed
	s.Ctx = s.App.NewContext(true, tmproto.Header{Time: time.Now()})
	tokenfactoryModuleAccount := s.App.Keepers.AccountKeeper.GetAccount(s.Ctx, s.App.Keepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().Nil(tokenfactoryModuleAccount)

	// create module account
	s.App.Keepers.TokenFactoryKeeper.CreateModuleAccount(s.Ctx)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = s.App.Keepers.AccountKeeper.GetAccount(s.Ctx, s.App.Keepers.AccountKeeper.GetModuleAddress(types.ModuleName))
	s.Require().NotNil(tokenfactoryModuleAccount)

	// check that the account number of the module account is now initialized correctly
	s.Require().Equal(nextAccountNumber+1, tokenfactoryModuleAccount.GetAccountNumber())
}

func (s *KeeperTestSuite) TestBurnFromModuleAccount() {
	// Create Msg Server
	msgServer := keeper.NewMsgServerImpl(s.App.Keepers.TokenFactoryKeeper)

	// Create token factory token
	res, err := msgServer.CreateDenom(s.Ctx, &types.MsgCreateDenom{
		Sender:   s.TestAccs[0].String(),
		Subdenom: "bitcoin",
	})
	s.Require().NoError(err)

	denom := res.GetNewTokenDenom()

	// Gov address
	govAddr := s.App.Keepers.AccountKeeper.GetModuleAddress("gov")
	s.App.Keepers.AccountKeeper.SetModuleAccount(s.Ctx, authtypes.NewModuleAccount(authtypes.NewBaseAccount(govAddr, nil, 0, 0), "gov", authtypes.Minter))
	ma := s.App.Keepers.AccountKeeper.GetAccount(s.Ctx, govAddr)
	_, ok := ma.(authtypes.ModuleAccountI)
	s.Require().True(ok)

	_, err = msgServer.Mint(s.Ctx, &types.MsgMint{
		Sender:        s.TestAccs[0].String(),
		Amount:        sdk.NewCoin(denom, sdk.NewInt(1000)),
		MintToAddress: s.TestAccs[0].String(),
	})
	s.Require().NoError(err)

	// Send to gov address
	err = s.App.Keepers.BankKeeper.SendCoinsFromAccountToModule(s.Ctx, s.TestAccs[0], "gov", sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000))))
	s.Require().NoError(err)

	_, err = msgServer.Burn(s.Ctx, &types.MsgBurn{
		Sender:          s.TestAccs[0].String(),
		Amount:          sdk.NewCoin(denom, sdk.NewInt(1000)),
		BurnFromAddress: govAddr.String(),
	})

	require.Error(s.T(), err)
}
