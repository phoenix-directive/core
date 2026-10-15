package interchaintest

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	interchaintestrelayer "github.com/cosmos/interchaintest/v10/relayer"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

type ForwardMetadata struct {
	Receiver       string        `json:"receiver"`
	Port           string        `json:"port"`
	Channel        string        `json:"channel"`
	Timeout        time.Duration `json:"timeout"`
	Retries        *uint8        `json:"retries,omitempty"`
	Next           *string       `json:"next,omitempty"`
	RefundSequence *uint64       `json:"refund_sequence,omitempty"`
}

// TestPacketForwardMiddlewareRouter ensures the PFM module is set up properly and works as expected.
func TestPMF(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		ctx                                        = context.Background()
		client, network                            = interchaintest.DockerSetup(t)
		rep                                        = testreporter.NewNopReporter()
		eRep                                       = rep.RelayerExecReporter(t)
		chainID_A, chainID_B, chainID_C, chainID_D = "chain-a", "chain-b", "chain-c", "chain-d"
		chainA, chainB, chainC, chainD             *cosmos.CosmosChain
	)

	baseCfg := config
	// Leave enough time for the proposal/query/vote Docker commands on slower hosts.
	genesis := append([]cosmos.GenesisKV{}, defaultGenesisKV...)
	genesis = append(genesis, cosmos.GenesisKV{Key: "app_state.gov.params.voting_period", Value: "60s"})
	baseCfg.ModifyGenesis = cosmos.ModifyGenesis(genesis)
	// Allow testing a freshly built image without replacing the default image.
	if tag := os.Getenv("PFM_IMAGE_TAG"); tag != "" {
		image := Image
		image.Version = tag
		baseCfg.Images = []ibc.DockerImage{image}
	}

	// Set specific chain ids for each so they are their own unique networks
	baseCfg.ChainID = chainID_A
	configA := baseCfg

	baseCfg.ChainID = chainID_B
	configB := baseCfg

	baseCfg.ChainID = chainID_C
	configC := baseCfg

	baseCfg.ChainID = chainID_D
	configD := baseCfg

	// Create chain factory with multiple Juno individual networks.
	numVals := 1
	numFullNodes := 0

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "juno",
			ChainConfig:   configA,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "juno",
			ChainConfig:   configB,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "juno",
			ChainConfig:   configC,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "juno",
			ChainConfig:   configD,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	// Get chains from the chain factory
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chainA, chainB, chainC, chainD = chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain), chains[2].(*cosmos.CosmosChain), chains[3].(*cosmos.CosmosChain)

	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.CosmosRly,
		zaptest.NewLogger(t),
		// interchaintestrelayer.CustomDockerImage(IBCRelayerImage, IBCRelayerVersion, "100:1000"),
		interchaintestrelayer.StartupFlags("--processor", "events", "--block-history", "100"),
	).Build(t, client, network)

	const pathAB = "ab"
	const pathBC = "bc"
	const pathCD = "cd"

	ic := interchaintest.NewInterchain().
		AddChain(chainA).
		AddChain(chainB).
		AddChain(chainC).
		AddChain(chainD).
		AddRelayer(r, "relayer").
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainA,
			Chain2:  chainB,
			Relayer: r,
			Path:    pathAB,
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainB,
			Chain2:  chainC,
			Relayer: r,
			Path:    pathBC,
		}).
		AddLink(interchaintest.InterchainLink{
			Chain1:  chainC,
			Chain2:  chainD,
			Relayer: r,
			Path:    pathCD,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	userFunds := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, chainA, chainB, chainC, chainD)

	abChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainID_A, chainID_B)
	require.NoError(t, err)

	baChan := abChan.Counterparty

	cbChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainID_C, chainID_B)
	require.NoError(t, err)

	bcChan := cbChan.Counterparty

	dcChan, err := ibc.GetTransferChannel(ctx, r, eRep, chainID_D, chainID_C)
	require.NoError(t, err)

	cdChan := dcChan.Counterparty

	// Start the relayer on all paths
	err = r.StartRelayer(ctx, eRep, pathAB, pathBC, pathCD)
	require.NoError(t, err)

	t.Cleanup(
		func() {
			err := r.StopRelayer(ctx, eRep)
			if err != nil {
				t.Logf("an error occurred while stopping the relayer: %s", err)
			}
		},
	)

	// Get original account balances
	userA, userB, userC, userD := users[0], users[1], users[2], users[3]

	transferAmount := math.NewInt(100000)

	// Compose the prefixed denoms and ibc denom for asserting balances
	firstHopDenom := transfertypes.GetPrefixedDenom(baChan.PortID, baChan.ChannelID, chainA.Config().Denom)
	secondHopDenom := transfertypes.GetPrefixedDenom(cbChan.PortID, cbChan.ChannelID, firstHopDenom)
	thirdHopDenom := transfertypes.GetPrefixedDenom(dcChan.PortID, dcChan.ChannelID, secondHopDenom)

	firstHopDenomTrace := transfertypes.ParseDenomTrace(firstHopDenom)
	secondHopDenomTrace := transfertypes.ParseDenomTrace(secondHopDenom)
	thirdHopDenomTrace := transfertypes.ParseDenomTrace(thirdHopDenom)

	firstHopIBCDenom := firstHopDenomTrace.IBCDenom()
	secondHopIBCDenom := secondHopDenomTrace.IBCDenom()
	thirdHopIBCDenom := thirdHopDenomTrace.IBCDenom()

	firstHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainA.Config().Bech32Prefix, transfertypes.GetEscrowAddress(abChan.PortID, abChan.ChannelID))
	secondHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainB.Config().Bech32Prefix, transfertypes.GetEscrowAddress(bcChan.PortID, bcChan.ChannelID))
	thirdHopEscrowAccount := sdk.MustBech32ifyAddressBytes(chainC.Config().Bech32Prefix, transfertypes.GetEscrowAddress(cdChan.PortID, cdChan.ChannelID))

	t.Run("multi-hop a->b->c->d", func(t *testing.T) {
		// Send packet from Chain A->Chain B->Chain C->Chain D

		transfer := ibc.WalletAmount{
			Address: userB.FormattedAddress(),
			Denom:   chainA.Config().Denom,
			Amount:  transferAmount,
		}

		secondHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userD.FormattedAddress(),
				Channel:  cdChan.ChannelID,
				Port:     cdChan.PortID,
			},
		}
		nextBz, err := json.Marshal(secondHopMetadata)
		require.NoError(t, err)
		next := string(nextBz)

		firstHopMetadata := &PacketMetadata{
			Forward: &ForwardMetadata{
				Receiver: userC.FormattedAddress(),
				Channel:  bcChan.ChannelID,
				Port:     bcChan.PortID,
				Next:     &next,
			},
		}

		memo, err := json.Marshal(firstHopMetadata)
		require.NoError(t, err)

		chainAHeight, err := chainA.Height(ctx)
		require.NoError(t, err)

		transferTx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), transfer, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		_, err = testutil.PollForAck(ctx, chainA, chainAHeight, chainAHeight+45, transferTx.Packet)
		require.NoError(t, err)
		err = testutil.WaitForBlocks(ctx, 2, chainA)
		require.NoError(t, err)

		chainABalance, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
		require.NoError(t, err)

		chainBBalance, err := chainB.GetBalance(ctx, userB.FormattedAddress(), firstHopIBCDenom)
		require.NoError(t, err)

		chainCBalance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
		require.NoError(t, err)

		chainDBalance, err := chainD.GetBalance(ctx, userD.FormattedAddress(), thirdHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, userFunds.Sub(transferAmount), chainABalance)
		require.Equal(t, math.ZeroInt(), chainBBalance)
		require.Equal(t, math.ZeroInt(), chainCBalance)
		require.Equal(t, transferAmount, chainDBalance)

		firstHopEscrowBalance, err := chainA.GetBalance(ctx, firstHopEscrowAccount, chainA.Config().Denom)
		require.NoError(t, err)

		secondHopEscrowBalance, err := chainB.GetBalance(ctx, secondHopEscrowAccount, firstHopIBCDenom)
		require.NoError(t, err)

		thirdHopEscrowBalance, err := chainC.GetBalance(ctx, thirdHopEscrowAccount, secondHopIBCDenom)
		require.NoError(t, err)

		require.Equal(t, transferAmount, firstHopEscrowBalance)
		require.Equal(t, transferAmount, secondHopEscrowBalance)
		require.Equal(t, transferAmount, thirdHopEscrowBalance)
	})
	if t.Failed() {
		return
	}

	// The first transfer seeds B with A's voucher. Its supply is stable, unlike
	// native uluna, so the quota snapshot and the expected flow are deterministic.
	authority, err := chainB.GetGovernanceAddress(ctx)
	require.NoError(t, err)
	var messages []json.RawMessage
	for _, limit := range []struct {
		channel, send, recv string
	}{
		{baChan.ChannelID, "10", "10"},
		{bcChan.ChannelID, "5", "10"},
	} {
		message, err := json.Marshal(map[string]string{
			"@type": "/ratelimit.v1.MsgAddRateLimit", "authority": authority,
			"denom": firstHopIBCDenom, "channel_or_client_id": limit.channel,
			"max_percent_send": limit.send, "max_percent_recv": limit.recv,
			"duration_hours": "24",
		})
		require.NoError(t, err)
		messages = append(messages, message)
	}
	proposal, err := chainB.SubmitProposal(ctx, userB.KeyName(), cosmos.TxProposalv1{
		Messages: messages, Deposit: "550000000" + chainB.Config().Denom,
		Title: "Meter forwarded transfers", Summary: "Exercise PFM rate limit accounting",
	})
	require.NoError(t, err)
	proposalID, err := strconv.ParseUint(proposal.ProposalID, 10, 64)
	require.NoError(t, err)
	require.NoError(t, chainB.VoteOnProposalAllValidators(ctx, proposalID, cosmos.ProposalVoteYes))
	_, err = cosmos.PollForProposalStatusV1(ctx, chainB, proposal.Height, proposal.Height+45, proposalID, govv1.StatusPassed)
	require.NoError(t, err)

	type flow struct {
		Inflow       math.Int `json:"inflow"`
		Outflow      math.Int `json:"outflow"`
		ChannelValue math.Int `json:"channel_value"`
	}
	queryFlow := func(t *testing.T, channel string) flow {
		t.Helper()
		response, _, err := chainB.GetNode().ExecQuery(ctx, "ratelimit", "rate-limit", channel, "--denom", firstHopIBCDenom)
		require.NoError(t, err)
		var limit struct {
			Flow flow `json:"flow"`
		}
		require.NoError(t, json.Unmarshal(response, &limit))
		require.True(t, limit.Flow.ChannelValue.Equal(transferAmount), "unexpected quota snapshot: %s", response)
		return limit.Flow
	}
	assertFlows := func(t *testing.T, amount math.Int) {
		t.Helper()
		incoming := queryFlow(t, baChan.ChannelID)
		outgoing := queryFlow(t, bcChan.ChannelID)
		require.True(t, incoming.Inflow.Equal(amount), "incoming flow: %s != %s", incoming.Inflow, amount)
		require.True(t, incoming.Outflow.IsZero())
		require.True(t, outgoing.Outflow.Equal(amount), "outgoing flow: %s != %s", outgoing.Outflow, amount)
		require.True(t, outgoing.Inflow.IsZero())
	}
	sendForward := func(t *testing.T, receiver string, amount math.Int, timeout time.Duration, retries *uint8) (ibc.Tx, int64) {
		t.Helper()
		memo, err := json.Marshal(PacketMetadata{Forward: &ForwardMetadata{
			Receiver: receiver, Port: bcChan.PortID, Channel: bcChan.ChannelID,
			Timeout: timeout, Retries: retries,
		}})
		require.NoError(t, err)
		height, err := chainA.Height(ctx)
		require.NoError(t, err)
		tx, err := chainA.SendIBCTransfer(ctx, abChan.ChannelID, userA.KeyName(), ibc.WalletAmount{
			Address: userB.FormattedAddress(), Denom: chainA.Config().Denom, Amount: amount,
		}, ibc.TransferOptions{Memo: string(memo)})
		require.NoError(t, err)
		return tx, height
	}
	waitAck := func(t *testing.T, tx ibc.Tx, height int64, success bool) {
		t.Helper()
		ack, err := testutil.PollForAck(ctx, chainA, height, height+60, tx.Packet)
		require.NoError(t, err)
		// PollForAck finds the message in a block before its application state is
		// necessarily committed. Wait before asserting refunds and flow counters.
		require.NoError(t, testutil.WaitForBlocks(ctx, 2, chainA, chainB))
		var response struct {
			Result string `json:"result"`
			Error  string `json:"error"`
		}
		require.NoError(t, json.Unmarshal(ack.Acknowledgement, &response))
		if success {
			require.Empty(t, response.Error)
			require.NotEmpty(t, response.Result)
		} else {
			require.NotEmpty(t, response.Error)
		}
	}

	forwardAmount := transferAmount.QuoRaw(100) // 1% of the quota snapshot
	t.Run("forwarded success is counted on both B channels", func(t *testing.T) {
		tx, height := sendForward(t, userC.FormattedAddress(), forwardAmount, time.Minute, nil)
		waitAck(t, tx, height, true)
		assertFlows(t, forwardAmount)
		balance, err := chainC.GetBalance(ctx, userC.FormattedAddress(), secondHopIBCDenom)
		require.NoError(t, err)
		require.True(t, balance.Equal(forwardAmount))
	})
	if t.Failed() {
		return
	}

	assertRefund := func(t *testing.T, send func() (ibc.Tx, int64)) {
		t.Helper()
		before, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
		require.NoError(t, err)
		tx, height := send()
		waitAck(t, tx, height, false)
		after, err := chainA.GetBalance(ctx, userA.FormattedAddress(), chainA.Config().Denom)
		require.NoError(t, err)
		require.True(t, before.Equal(after), "source funds were not refunded")
		assertFlows(t, forwardAmount)
	}
	t.Run("destination error undoes forwarded inflow and outflow", func(t *testing.T) {
		assertRefund(t, func() (ibc.Tx, int64) {
			return sendForward(t, "invalid-receiver", forwardAmount, time.Minute, nil)
		})
	})
	if t.Failed() {
		return
	}
	t.Run("terminal forward timeout undoes inflow and outflow", func(t *testing.T) {
		// Relay A->B but leave B->C unrelayed until its short timeout expires.
		require.NoError(t, r.StopRelayer(ctx, eRep))
		require.NoError(t, r.StartRelayer(ctx, eRep, pathAB))
		assertRefund(t, func() (ibc.Tx, int64) {
			noRetries := uint8(0)
			tx, height := sendForward(t, userC.FormattedAddress(), forwardAmount, 5*time.Second, &noRetries)
			require.NoError(t, testutil.WaitForBlocks(ctx, 8, chainA, chainB, chainC))
			// The receive is asynchronous: retain its accounting until the final ack.
			assertFlows(t, forwardAmount.MulRaw(2))
			require.NoError(t, r.StopRelayer(ctx, eRep))
			require.NoError(t, r.StartRelayer(ctx, eRep, pathAB, pathBC, pathCD))
			return tx, height
		})
	})
	if t.Failed() {
		return
	}
	t.Run("forward cannot exceed B outbound quota", func(t *testing.T) {
		assertRefund(t, func() (ibc.Tx, int64) {
			// 1% already sent + 5% exceeds B->C's 5% send limit, but not A->B's 10% receive limit.
			return sendForward(t, userC.FormattedAddress(), forwardAmount.MulRaw(5), time.Minute, nil)
		})
	})
	t.Run("forward cannot exceed B inbound quota", func(t *testing.T) {
		assertRefund(t, func() (ibc.Tx, int64) {
			return sendForward(t, userC.FormattedAddress(), transferAmount.QuoRaw(10), time.Minute, nil)
		})
	})
}
