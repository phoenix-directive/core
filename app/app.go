package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	tmos "github.com/cometbft/cometbft/libs/os"

	"github.com/terra-money/core/v2/app/custom_queriers"
	"github.com/terra-money/core/v2/app/keepers"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"

	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	cosmosante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	allianceclient "github.com/terra-money/alliance/x/alliance/client"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	// unnamed import of statik for swagger UI support
	_ "github.com/terra-money/core/v2/client/docs/statik"

	"github.com/terra-money/core/v2/app/ante"
	terraappconfig "github.com/terra-money/core/v2/app/config"
	terraappparams "github.com/terra-money/core/v2/app/params"
	"github.com/terra-money/core/v2/app/post"
	"github.com/terra-money/core/v2/app/rpc"
)

var (
	_ servertypes.Application = (*TerraApp)(nil)
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+terraappconfig.AppName)
}

// TerraApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type TerraApp struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	Keepers keepers.TerraAppKeepers

	invCheckPeriod uint

	// the module manager
	mm           *module.Manager
	basicManager module.BasicManager
	// the configurator
	configurator module.Configurator
}

// NewTerraApp returns a reference to an initialized Terra.
func NewTerraApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig terraappparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmConfig wasmtypes.NodeConfig,
	baseAppOptions ...func(*baseapp.BaseApp),
) *TerraApp {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(terraappconfig.AppName, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())
	app := &TerraApp{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}
	app.Keepers = keepers.NewTerraAppKeepers(
		appCodec,
		bApp,
		cdc,
		appOpts,
		app.GetWasmOpts(appOpts),
		homePath,
		logger,
	)
	app.keys = app.Keepers.GetKVStoreKey()
	app.tkeys = app.Keepers.GetTransientStoreKey()
	app.memKeys = app.Keepers.GetMemoryStoreKey()
	bApp.SetParamStore(&app.Keepers.ConsensusParamsKeeper.ParamsStore)

	// TODO: Verify if this is needed
	baseAppOptions = append(baseAppOptions, baseapp.SetOptimisticExecution())

	// upgrade handlers
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	app.mm = module.NewManager(appModules(app, encodingConfig, skipGenesisInvariants)...)

	// NOTE: Any module instantiated in the module manager that is later modified
	app.mm.SetOrderPreBlockers(preBlockersOrder...)
	// must be passed by reference here.
	app.mm.SetOrderBeginBlockers(beginBlockersOrder...)

	app.mm.SetOrderEndBlockers(endBlockerOrder...)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(initGenesisOrder...)

	app.mm.RegisterInvariants(&app.Keepers.CrisisKeeper)
	app.mm.RegisterServices(app.configurator)

	// initialize stores
	app.MountKVStores(app.keys)
	app.MountTransientStores(app.tkeys)
	app.MountMemoryStores(app.memKeys)

	// register upgrade
	app.RegisterUpgradeHandlers()
	app.RegisterUpgradeStores()
	txCounterStoreService := runtime.NewKVStoreService(app.keys[wasmtypes.StoreKey])

	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			HandlerOptions: cosmosante.HandlerOptions{
				AccountKeeper:   app.Keepers.AccountKeeper,
				BankKeeper:      app.Keepers.BankKeeper,
				FeegrantKeeper:  app.Keepers.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  cosmosante.DefaultSigVerificationGasConsumer,
			},
			BankKeeper:            app.Keepers.BankKeeper,
			FeeShareKeeper:        app.Keepers.FeeShareKeeper,
			IBCkeeper:             app.Keepers.IBCKeeper,
			TXCounterStoreService: txCounterStoreService,
			NodeConfig:            wasmConfig,
		},
	)
	if err != nil {
		panic(err)
	}
	postHandler := post.NewPostHandler(
		post.HandlerOptions{
			FeeShareKeeper: app.Keepers.FeeShareKeeper,
			BankKeeper:     app.Keepers.BankKeeper,
			WasmKeeper:     app.Keepers.WasmKeeper,
		},
	)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(anteHandler)
	app.SetPostHandler(postHandler)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		// Initialize and seal the capability keeper so all persistent capabilities
		// are loaded in-memory and prevent any further modules from creating scoped
		// sub-keepers.
		// This must be done during creation of baseapp rather than in InitChain so
		// that in-memory capabilities get regenerated on app restart.
		// Note that since this reads from the store, we can only perform it when
		// `loadLatest` is set to true.
		app.Keepers.CapabilityKeeper.Seal()
	}

	return app
}

// // Commit implements types.Application.
// // Subtle: this method shadows the method (*BaseApp).Commit of TerraApp.BaseApp.
// func (app *TerraApp) Commit() (*abcitypes.ResponseCommit, error) {
// 	panic("unimplemented")
// }

// // FinalizeBlock implements types.Application.
// // Subtle: this method shadows the method (*BaseApp).FinalizeBlock of TerraApp.BaseApp.
// func (app *TerraApp) FinalizeBlock(*abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
// 	panic("unimplemented")
// }

// // PrepareProposal implements types.Application.
// // Subtle: this method shadows the method (*BaseApp).PrepareProposal of TerraApp.BaseApp.
// func (app *TerraApp) PrepareProposal(*abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
// 	panic("unimplemented")
// }

// // ProcessProposal implements types.Application.
// // Subtle: this method shadows the method (*BaseApp).ProcessProposal of TerraApp.BaseApp.
// func (app *TerraApp) ProcessProposal(*abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
// 	panic("unimplemented")
// }

// RegisterAPIRoutes implements types.Application.
func (app *TerraApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	rpc.RegisterHealthcheckRoute(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register legacy and grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// // RegisterGRPCServer implements types.Application.
// // Subtle: this method shadows the method (*BaseApp).RegisterGRPCServer of TerraApp.BaseApp.
// func (app *TerraApp) RegisterGRPCServer(grpc.Server) {
// 	panic("unimplemented")
// }

// RegisterNodeService implements types.Application.
func (app *TerraApp) RegisterNodeService(clientCtx client.Context, config config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), config)
}

// RegisterTendermintService implements types.Application.
func (app *TerraApp) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterTxService implements types.Application.
func (app *TerraApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// ----------------------------
// Extended Functions
// ----------------------------
func (app *TerraApp) GetWasmOpts(appOpts servertypes.AppOptions) []wasmkeeper.Option {
	var wasmOpts []wasmkeeper.Option
	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	wasmOpts = append(wasmOpts, custom_queriers.RegisterCustomPlugins(
		&app.Keepers.BankKeeper.BaseKeeper,
		&app.Keepers.TokenFactoryKeeper,
		&app.Keepers.AllianceKeeper)...,
	)

	return wasmOpts
}

// InitChainer application update at chain initialization
func (app *TerraApp) InitChainer(ctx sdk.Context, req *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.Keepers.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// PreBlocker application updates every pre block
func (app *TerraApp) PreBlocker(ctx sdk.Context, _ *abcitypes.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *TerraApp) BeginBlocker(ctx sdk.Context) (sdk.BeginBlock, error) {
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *TerraApp) EndBlocker(ctx sdk.Context) (sdk.EndBlock, error) {
	return app.mm.EndBlock(ctx)
}

// LoadHeight loads a particular height
func (app *TerraApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		allianceclient.CreateAllianceProposalHandler,
		allianceclient.UpdateAllianceProposalHandler,
		allianceclient.DeleteAllianceProposalHandler,
	)

	return govProposalHandlers
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *TerraApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, found := app.Keepers.ParamsKeeper.GetSubspace(moduleName)
	if !found {
		panic("Module with '" + moduleName + "' name does not exist")
	}

	return subspace
}

func (app TerraApp) GetConfigurator() module.Configurator {
	return app.configurator
}

func (app TerraApp) GetModuleManager() *module.Manager {
	return app.mm
}

func (app TerraApp) GetAppCodec() codec.Codec {
	return app.appCodec
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *TerraApp) DefaultGenesis() map[string]json.RawMessage {
	return a.basicManager.DefaultGenesis(a.appCodec)
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *TerraApp) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

func (app *TerraApp) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		AddressCodec:          addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}
