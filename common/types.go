package common

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// TOOD: add json tags
type SubmitTransactionRequest struct {
	From      ethcommon.Address
	To        ethcommon.Address
	Value     *big.Int
	Gas       *big.Int
	Nonce     *big.Int
	Data      []byte
	Signature []byte
}

type SubmitTransactionResponse struct {
	TxHash ethcommon.Hash `json:"transactionHash"`
}

type Forwarder interface {
	GetNonce(opts *bind.CallOpts, from ethcommon.Address) (*big.Int, error)
	Verify(opts *bind.CallOpts, req ForwardRequest, signature []byte) (bool, error)
	Execute(opts *bind.TransactOpts, req ForwardRequest, signature []byte) (*types.Transaction, error)
}

type ForwardRequest interface {
	FromSubmitTransactionRequest(*SubmitTransactionRequest)
	Pack() ([]byte, error)
}