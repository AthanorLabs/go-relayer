package relayer

import (
	"context"

	"github.com/athanorlabs/go-relayer/common"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("relayer")

// NewForwardRequestFunc returns a new, empty ForwardRequest that must be
// compatible with the Forwarder the relayer is configured to use.
type NewForwardRequestFunc func() common.ForwardRequest

// ValidateTransactionFunc is a user-set transaction validation function.
// For example, the user may wish to validate the contract being called,
// or the calldata, to only allow forward requests to specific contracts
// or with specific data.
type ValidateTransactionFunc func(*common.SubmitTransactionRequest) error

// Relayer represents a transaction relayer.
// It contains an Ethereum client and a private key, allowing it to forward transactions.
type Relayer struct {
	ctx                     context.Context
	ec                      *ethclient.Client
	forwarder               common.Forwarder
	callOpts                *bind.CallOpts
	txOpts                  *bind.TransactOpts
	newForwardRequestFunc   NewForwardRequestFunc
	validateTransactionFunc ValidateTransactionFunc
}

// Config ...
type Config struct {
	Ctx                     context.Context
	EthClient               *ethclient.Client
	Forwarder               common.Forwarder
	Key                     *common.Key
	NewForwardRequestFunc   NewForwardRequestFunc
	ValidateTransactionFunc ValidateTransactionFunc
}

func NewRelayer(cfg *Config) (*Relayer, error) {
	if cfg.Forwarder == nil {
		return nil, errMustSetForwarder
	}

	if cfg.NewForwardRequestFunc == nil {
		return nil, errMustSetNewForwardRequestFunc
	}

	if cfg.ValidateTransactionFunc == nil {
		return nil, errMustSetValidateTransactionFunc
	}

	if cfg.Key == nil {
		return nil, errMustSetKey
	}

	if cfg.EthClient == nil {
		return nil, errMustSetEthClient
	}

	callOpts := &bind.CallOpts{
		From:    cfg.Key.Address(),
		Context: cfg.Ctx,
	}

	chainID, err := cfg.EthClient.ChainID(cfg.Ctx)
	if err != nil {
		return nil, err
	}

	txOpts, err := bind.NewKeyedTransactorWithChainID(cfg.Key.PrivateKey(), chainID)
	if err != nil {
		return nil, err
	}

	return &Relayer{
		ctx:                     cfg.Ctx,
		ec:                      cfg.EthClient,
		forwarder:               cfg.Forwarder,
		callOpts:                callOpts,
		txOpts:                  txOpts,
		newForwardRequestFunc:   cfg.NewForwardRequestFunc,
		validateTransactionFunc: cfg.ValidateTransactionFunc,
	}, nil
}

func (s *Relayer) SubmitTransaction(req *common.SubmitTransactionRequest) (*common.SubmitTransactionResponse, error) {
	err := validateRequest(req)
	if err != nil {
		return nil, err
	}

	err = s.validateTransactionFunc(req)
	if err != nil {
		return nil, err
	}

	fwdReq := s.newForwardRequestFunc()
	fwdReq.FromSubmitTransactionRequest(req)

	// verify sig beforehand
	ok, err := s.forwarder.Verify(
		s.callOpts,
		fwdReq,
		req.DomainSeparator,
		req.RequestTypeHash,
		req.SuffixData,
		req.Signature,
	)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errFailedToVerify
	}

	tx, err := s.forwarder.Execute(
		s.txOpts,
		fwdReq,
		req.DomainSeparator,
		req.RequestTypeHash,
		req.SuffixData,
		req.Signature,
	)
	if err != nil {
		return nil, err
	}

	log.Infof("submitted transaction %s", tx.Hash())
	return &common.SubmitTransactionResponse{
		TxHash: tx.Hash(),
	}, nil
}

func validateRequest(req *common.SubmitTransactionRequest) error {
	if req.From == (ethcommon.Address{}) {
		return errFromNotProvided
	}

	if req.Signature == nil {
		return errSignatureNotProvided
	}

	if req.Gas == nil {
		return errGasNotProvided
	}

	if req.Nonce == nil {
		return errNonceNotProvided
	}

	return nil
}
