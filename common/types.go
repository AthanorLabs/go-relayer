package common

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type SubmitTransactionRequest struct {
	From      ethcommon.Address `json:"from"`
	To        ethcommon.Address `json:"to"`
	Value     *big.Int          `json:"value"`
	Gas       *big.Int          `json:"gas"`
	Nonce     *big.Int          `json:"nonce"`
	Data      []byte            `json:"data"`
	Signature []byte            `json:"signature"`

	// GSN-specific
	ValidUntilTime *big.Int `json:"validUntilTime,omitempty"`

	DomainSeparator [32]byte `json:"domainSeparator,omitempty"`
	RequestTypeHash [32]byte `json:"requestTypeHash,omitempty"`
	SuffixData      []byte   `json:"suffixData,omitempty"`
}

type SubmitTransactionResponse struct {
	TxHash ethcommon.Hash `json:"transactionHash"`
}

type Forwarder interface {
	GetNonce(opts *bind.CallOpts, from ethcommon.Address) (*big.Int, error)

	Verify(
		opts *bind.CallOpts,
		req ForwardRequest,
		domainSeparator,
		requestTypeHash [32]byte,
		suffixData,
		signature []byte,
	) (bool, error)

	Execute(
		opts *bind.TransactOpts,
		req ForwardRequest,
		domainSeparator,
		requestTypeHash [32]byte,
		suffixData,
		signature []byte,
	) (*types.Transaction, error)
}

type ForwardRequest interface {
	// FromSubmitTransactionRequest set the type underlying the ForwardRequest
	// using a *SubmitTransactionRequest.
	//
	// Note: not all fields in the *SubmitTransactionRequest need be used depending
	// on the implementation.
	FromSubmitTransactionRequest(*SubmitTransactionRequest)

	// Pack uses ABI encoding to pack the underlying ForwardRequest, appending
	// optional `suffixData` to the end.
	//
	// See examples/gsn_forwarder/IForwarderForwardRequest.Pack() or
	// examples/minimal_forwarder/IMinimalForwarderForwardRequest.Pack()
	// for details.
	Pack(suffixData []byte) ([]byte, error)
}
