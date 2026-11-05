package app

import (
	"encoding/json"

	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icagenesistypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/genesis/types"
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	"github.com/terra-money/core/v2/app/config"
	tokenfactorytypes "github.com/terra-money/core/v2/x/tokenfactory/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"cosmossdk.io/math"
)

// GenesisState - The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	return ModuleBasics.DefaultGenesis(cdc)
}

// SetDefaultTerraConfig generates the default state for Terra's Core application.
func (genState GenesisState) SetDefaultTerraConfig(cdc codec.JSONCodec) GenesisState {
	// customize bond denom
	var stakingGenState stakingtypes.GenesisState
	cdc.MustUnmarshalJSON(genState[stakingtypes.ModuleName], &stakingGenState)
	stakingGenState.Params.BondDenom = config.BondDenom
	genState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(&stakingGenState)

	var crisisGenState crisistypes.GenesisState
	cdc.MustUnmarshalJSON(genState[crisistypes.ModuleName], &crisisGenState)
	crisisGenState.ConstantFee.Denom = config.BondDenom
	genState[crisistypes.ModuleName] = cdc.MustMarshalJSON(&crisisGenState)

	var govGenState govtypesv1.GenesisState
	cdc.MustUnmarshalJSON(genState[govtypes.ModuleName], &govGenState)
	govGenState.Params.MinDeposit[0].Denom = config.BondDenom
	genState[govtypes.ModuleName] = cdc.MustMarshalJSON(&govGenState)

	var mintGenState minttypes.GenesisState
	cdc.MustUnmarshalJSON(genState[minttypes.ModuleName], &mintGenState)
	mintGenState.Params.MintDenom = config.BondDenom
	genState[minttypes.ModuleName] = cdc.MustMarshalJSON(&mintGenState)

	var tokenFactoryGenState tokenfactorytypes.GenesisState
	cdc.MustUnmarshalJSON(genState[tokenfactorytypes.ModuleName], &tokenFactoryGenState)
	tokenFactoryGenState.Params.DenomCreationFee = sdk.NewCoins(sdk.NewCoin(config.BondDenom, math.NewInt(10000000)))
	genState[tokenfactorytypes.ModuleName] = cdc.MustMarshalJSON(&tokenFactoryGenState)

	var icaGenState icagenesistypes.GenesisState
	cdc.MustUnmarshalJSON(genState[icatypes.ModuleName], &icaGenState)
	icaGenState.ControllerGenesisState.Params = icacontrollertypes.Params{
		ControllerEnabled: true,
	}
	icaGenState.HostGenesisState.Params = icahosttypes.Params{
		HostEnabled:   true,
		AllowMessages: icaAllowedMsgs(),
	}
	genState[icatypes.ModuleName] = cdc.MustMarshalJSON(&icaGenState)

	return genState
}

func icaAllowedMsgs() []string {
	return []string{
		config.AuthzMsgExec,
		config.AuthzMsgGrant,
		config.AuthzMsgRevoke,
		config.BankMsgSend,
		config.BankMsgMultiSend,
		config.DistrMsgSetWithdrawAddr,
		config.DistrMsgWithdrawValidatorCommission,
		config.DistrMsgFundCommunityPool,
		config.DistrMsgWithdrawDelegatorReward,
		config.FeegrantMsgGrantAllowance,
		config.FeegrantMsgRevokeAllowance,
		config.GovMsgVoteWeighted,
		config.GovMsgSubmitProposal,
		config.GovMsgDeposit,
		config.GovMsgVote,
		config.StakingMsgEditValidator,
		config.StakingMsgDelegate,
		config.StakingMsgUndelegate,
		config.StakingMsgBeginRedelegate,
		config.StakingMsgCreateValidator,
		config.VestingMsgCreateVestingAccount,
		config.TransferMsgTransfer,
		config.WasmMsgStoreCode,
		config.WasmMsgInstantiateContract,
		config.WasmMsgExecuteContract,
		config.WasmMsgMigrateContract,
	}
}
