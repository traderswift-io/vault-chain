package main

import (
	"bytes"
	"context"
	"flag"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"vault-chain/chainprefix"

	cmtcfg "github.com/cometbft/cometbft/config"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cometbft/cometbft/proxy"
	"github.com/cometbft/cometbft/rpc/client/local"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var localChainDirectory = flag.String("local-chain-directory", "", "isolated runner directory for the local consensus scenario")
var localChainP2PPort = flag.Int("local-chain-p2p-port", 0, "reserved loopback P2P port selected by the runner")
var localChainRPCPort = flag.Int("local-chain-rpc-port", 0, "reserved loopback RPC port selected by the runner")

func TestLocalRoundTrip(t *testing.T) {
	if *localChainDirectory == "" {
		t.Skip("runner invokes this scenario with an isolated directory")
	}
	chainprefix.Apply()
	root, err := filepath.Abs(*localChainDirectory)
	if err != nil || filepath.Base(root) != "roundtrip" || filepath.Base(filepath.Dir(root)) != "_build" {
		t.Fatal("expected isolated _build/roundtrip directory")
	}
	workingDirectory, err := filepath.Abs(".")
	if err != nil || root != filepath.Join(workingDirectory, "_build", "roundtrip") {
		t.Fatal("local chain directory must be inside the current tree")
	}
	if *localChainP2PPort < 47000 || *localChainP2PPort > 47999 || *localChainRPCPort < 47000 || *localChainRPCPort > 47999 || *localChainP2PPort == *localChainRPCPort {
		t.Fatal("two distinct ports in 47000-47999 are required")
	}
	configuration := cmtcfg.DefaultConfig()
	configuration.SetRoot(root)
	configuration.Moniker = "local-check"
	configuration.RPC.ListenAddress = "tcp://127.0.0.1:" + strconv.Itoa(*localChainRPCPort)
	configuration.P2P.ListenAddress = "tcp://127.0.0.1:" + strconv.Itoa(*localChainP2PPort)
	configuration.P2P.ExternalAddress = ""
	configuration.P2P.Seeds = ""
	configuration.P2P.PersistentPeers = ""
	configuration.P2P.PexReactor = false
	configuration.Consensus.CreateEmptyBlocks = false
	configuration.Instrumentation.Prometheus = false
	cmtcfg.EnsureRoot(root)
	validator := privval.LoadOrGenFilePV(configuration.PrivValidatorKeyFile(), configuration.PrivValidatorStateFile())
	public, err := validator.GetPubKey()
	if err != nil {
		t.Fatal(err)
	}
	transport, err := p2p.LoadOrGenNodeKey(configuration.NodeKeyFile())
	if err != nil {
		t.Fatal(err)
	}
	admin := ed25519.GenPrivKey()
	app, encoded := applicationFixture(t, admin, [][]byte{public.Bytes()})
	newOwner := ed25519.GenPrivKey()
	ownerAddress := sdk.AccAddress(newOwner.PubKey().Address())
	if bytes.Contains(encoded, []byte(ownerAddress.String())) {
		t.Fatal("round trip owner must be absent from genesis")
	}
	genesis := &cmttypes.GenesisDoc{GenesisTime: time.Now().UTC(), ChainID: chainprefix.ChainID, InitialHeight: 1, ConsensusParams: cmttypes.DefaultConsensusParams(), AppState: encoded}
	genesis.Validators = []cmttypes.GenesisValidator{{Address: public.Address(), PubKey: public, Power: 100, Name: "local-check"}}
	comet, err := node.NewNode(configuration, validator, transport, proxy.NewLocalClientCreator(server.NewCometABCIWrapper(app)), func() (*cmttypes.GenesisDoc, error) { return genesis, nil }, cmtcfg.DefaultDBProvider, node.DefaultMetricsProvider(configuration.Instrumentation), cmtlog.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	if err := comet.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := comet.Stop(); err != nil {
			t.Error(err)
		}
		comet.Wait()
	})
	client := local.New(comet)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	blocks, err := comet.EventBus().Subscribe(ctx, "roundtrip-blocks", cmttypes.EventQueryNewBlock, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer comet.EventBus().UnsubscribeAll(context.Background(), "roundtrip-blocks")
	initial, err := client.ABCIInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}
	committedHeight := initial.Response.LastBlockHeight
	for committedHeight < genesis.InitialHeight {
		select {
		case event, open := <-blocks.Out():
			if !open {
				t.Fatal("consensus event subscription closed before initial commit")
			}
			if block, ok := event.Data().(cmttypes.EventDataNewBlock); ok {
				committedHeight = block.Block.Height
			}
		case <-ctx.Done():
			t.Fatal("initial consensus commit deadline exceeded")
		}
	}
	t.Log("PASS initial consensus commit")
	sequence := uint64(0)
	broadcast := func(message sdk.Msg) []byte {
		t.Helper()
		transaction := signedTransactionNumber(t, app, newOwner, FirstContactAccountNumber(ownerAddress), sequence, message)
		result, err := client.BroadcastTxCommit(ctx, transaction)
		if err != nil {
			t.Fatal(err)
		}
		if result.CheckTx.Code != 0 || result.TxResult.Code != 0 {
			t.Fatalf("transaction refused: %+v", result)
		}
		sequence++
		for {
			select {
			case event, open := <-blocks.Out():
				if !open {
					t.Fatal("consensus event subscription closed")
				}
				block, ok := event.Data().(cmttypes.EventDataNewBlock)
				if ok && block.Block.Height >= result.Height {
					return encodeEvents(t, result.TxResult.Events, block)
				}
			case <-ctx.Done():
				t.Fatal("consensus event deadline exceeded")
			}
		}
	}
	if !strings.HasPrefix(configuration.RPC.ListenAddress, "tcp://127.0.0.1:") {
		t.Fatal("nonlocal listener")
	}
	runLocalMessages(t, app, newOwner, client, ctx, broadcast, root)
	t.Log("PASS fresh non-genesis owner round trip")
}
