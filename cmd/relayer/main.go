package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/go-relayer/common"
	contracts "github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/athanorlabs/go-relayer/relayer"
	"github.com/athanorlabs/go-relayer/rpc"

	logging "github.com/ipfs/go-log"
)

const (
	flagEndpoint         = "endpoint"
	flagForwarderAddress = "forwarder-address"
	flagKey              = "key"
	flagRPCPort          = "rpc-port"
	flagDeploy           = "deploy"
	flagLog              = "log-level"
)

var (
	log = logging.Logger("main")

	flags = []cli.Flag{
		&cli.StringFlag{
			Name:  flagEndpoint,
			Value: "http://localhost:8545",
			Usage: "Ethereum RPC endpoint",
		},
		&cli.StringFlag{
			Name:  flagKey,
			Value: "eth.key",
			Usage: "Path to file containing Ethereum private key",
		},
		&cli.UintFlag{
			Name:  flagRPCPort,
			Value: 7799,
			Usage: "Relayer RPC server port",
		},
		&cli.BoolFlag{
			Name:  flagDeploy,
			Usage: "Deploy an instance of the forwarder contract",
		},
		&cli.StringFlag{
			Name:  flagForwarderAddress,
			Usage: "Forwarder contract address",
		},
		&cli.StringFlag{
			Name:  flagLog,
			Value: "info",
			Usage: "Set log level: one of [error|warn|info|debug]",
		},
	}

	errInvalidAddress       = errors.New("invalid forwarder address")
	errNoEthereumPrivateKey = errors.New("must provide ethereum private key with --key")
)

func main() {
	app := &cli.App{
		Name:    "relayer",
		Usage:   "Ethereum transaction relayer",
		Version: getVersion(),
		Flags:   flags,
		Action:  run,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setLogLevels(c *cli.Context) error {
	const (
		levelError = "error"
		levelWarn  = "warn"
		levelInfo  = "info"
		levelDebug = "debug"
	)

	level := c.String(flagLog)
	switch level {
	case levelError, levelWarn, levelInfo, levelDebug:
	default:
		return fmt.Errorf("invalid log level %q", level)
	}

	_ = logging.SetLogLevel("main", level)
	_ = logging.SetLogLevel("relayer", level)
	_ = logging.SetLogLevel("rpc", level)
	return nil
}

func run(c *cli.Context) error {
	err := setLogLevels(c)
	if err != nil {
		return err
	}

	port := uint16(c.Uint(flagRPCPort))
	endpoint := c.String(flagEndpoint)
	ec, err := ethclient.Dial(endpoint)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(c.Context)
	go signalHandler(ctx, cancel)

	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return err
	}

	log.Infof("starting relayer with ethereum endpoint %s and chain ID %s", endpoint, chainID)

	key, err := getPrivateKey(c.String(flagKey))
	if err != nil {
		return err
	}

	contractAddr := c.String(flagForwarderAddress)
	deploy := c.Bool(flagDeploy)
	contractSet := contractAddr != ""
	if deploy && contractSet {
		return fmt.Errorf("flags --%s and --%s are mutually exclusive", flagDeploy, flagForwarderAddress)
	}
	if !deploy && !contractSet {
		return fmt.Errorf("either --%s or --%s is required", flagDeploy, flagForwarderAddress)
	}

	forwarder, err := deployOrGetForwarder(
		contractAddr,
		ec,
		key,
		chainID,
	)
	if err != nil {
		return err
	}

	cfg := &relayer.Config{
		Ctx:                   context.Background(),
		EthClient:             ec,
		Forwarder:             contracts.NewIForwarderWrapped(forwarder),
		Key:                   key,
		ChainID:               chainID,
		NewForwardRequestFunc: contracts.NewIForwarderForwardRequest,
	}

	r, err := relayer.NewRelayer(cfg)
	if err != nil {
		return err
	}

	rpcCfg := &rpc.Config{
		Ctx:     ctx,
		Address: fmt.Sprintf("127.0.0.1:%d", port),
		Relayer: r,
	}

	server, err := rpc.NewServer(rpcCfg)
	if err != nil {
		return err
	}

	err = server.Start()
	if errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func deployOrGetForwarder(
	addressString string,
	ec *ethclient.Client,
	key *common.Key,
	chainID *big.Int,
) (*contracts.IForwarder, error) { // TODO: change to interface
	txOpts, err := bind.NewKeyedTransactorWithChainID(key.PrivateKey(), chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to make transactor: %w", err)
	}

	if addressString == "" {
		address, tx, _, err := contracts.DeployForwarder(txOpts, ec)
		if err != nil {
			return nil, err
		}

		_, err = bind.WaitMined(context.Background(), ec, tx)
		if err != nil {
			return nil, err
		}

		log.Infof("deployed Forwarder.sol to %s", address)
		return contracts.NewIForwarder(address, ec)
	}

	ok := ethcommon.IsHexAddress(addressString)
	if !ok {
		return nil, errInvalidAddress
	}

	log.Infof("loaded Forwarder.sol at %s", ethcommon.HexToAddress(addressString))
	return contracts.NewIForwarder(ethcommon.HexToAddress(addressString), ec)
}

func getPrivateKey(keyFile string) (*common.Key, error) {
	if keyFile != "" {
		fileData, err := os.ReadFile(filepath.Clean(keyFile))
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		keyHex := strings.TrimSpace(string(fileData))
		return common.NewKeyFromPrivateKeyString(keyHex)
	}
	return nil, errNoEthereumPrivateKey
}
