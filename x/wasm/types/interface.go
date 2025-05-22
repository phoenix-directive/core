package types

import (
	"context"

	"github.com/CosmWasm/wasmd/x/wasm/types"
)

type KeeperInterface interface {
	AfterExecuteContract(ctx context.Context, msg *types.MsgExecuteContract, res *types.MsgExecuteContractResponse) error
}
