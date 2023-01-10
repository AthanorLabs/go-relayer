package net

import (
	"testing"

	"github.com/athanorlabs/go-relayer/common"

	"github.com/stretchr/testify/require"
)

func TestTransactionRequest(t *testing.T) {
	msg := &TransactionRequest{
		SubmitTransactionRequest: &common.SubmitTransactionRequest{},
	}

	enc, err := msg.Encode()
	require.NoError(t, err)

	dec, err := decodeMessage(enc)
	require.NoError(t, err)
	require.Equal(t, msg, dec)
}

func TestTransactionResponse(t *testing.T) {
	msg := &TransactionResponse{
		SubmitTransactionResponse: &common.SubmitTransactionResponse{},
	}

	enc, err := msg.Encode()
	require.NoError(t, err)

	dec, err := decodeMessage(enc)
	require.NoError(t, err)
	require.Equal(t, msg, dec)
}
