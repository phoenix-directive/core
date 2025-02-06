package keeper

import (
	"context"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	corestoretypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CosmWasm/wasmd/x/wasm/types"
	keepertypes "github.com/terra-money/core/v2/x/wasm/types"
)

var _ keepertypes.KeeperInterface = Keeper{}

type Keeper struct {
	*wasmkeeper.Keeper
	storeService corestoretypes.KVStoreService
	cdc          codec.Codec
}

func NewKeeper(
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistributionKeeper,
	ics4Wrapper types.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	capabilityKeeper types.CapabilityKeeper,
	portSource types.ICS20TransferPortSource,
	router wasmkeeper.MessageRouter,
	grpcQueryRouter wasmkeeper.GRPCQueryRouter,
	homeDir string,
	nodeConfig types.NodeConfig,
	vmConfig types.VMConfig,
	availableCapabilities []string,
	authority string,
	opts ...wasmkeeper.Option,
) Keeper {
	keeper := wasmkeeper.NewKeeper(
		cdc,
		storeService,
		accountKeeper,
		bankKeeper,
		stakingKeeper,
		distrKeeper,
		ics4Wrapper,
		channelKeeper,
		portKeeper,
		capabilityKeeper,
		portSource,
		router,
		grpcQueryRouter,
		homeDir,
		nodeConfig,
		vmConfig,
		availableCapabilities,
		authority,
		opts...,
	)

	return Keeper{
		Keeper:       &keeper,
		storeService: storeService,
		cdc:          cdc,
	}
}

// After executing the contract, get all executed
// contract addresses from the store, if there is
// a store already then check if the contract address
// exists in the list, if not then update the store,
// If the contract does not exist in the store, add it.
func (k Keeper) AfterExecuteContract(ctx context.Context, msg *types.MsgExecuteContract, res *types.MsgExecuteContractResponse) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	contracts, found := k.GetExecutedContractAddresses(sdkCtx)

	if found {
		for _, contract := range contracts.ContractAddresses {
			if contract == msg.Contract {
				return nil
			}
		}
	}

	contracts.ContractAddresses = append(contracts.ContractAddresses, msg.Contract)

	err := k.SetExecutedContractAddresses(sdkCtx, contracts)
	if err != nil {
		return err
	}
	return nil
}
