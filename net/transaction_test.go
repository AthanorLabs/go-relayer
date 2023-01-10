package net

import (
	"testing"

	"github.com/athanorlabs/go-relayer/common"

	"github.com/stretchr/testify/require"
)

func TestTransactionMessages(t *testing.T) {
	ha := newHost(t, basicTestConfig(t))
	hb := newHost(t, basicTestConfig(t))

	err := ha.h.Connect(ha.ctx, hb.AddrInfo())
	require.NoError(t, err)

	msg := &TransactionRequest{
		SubmitTransactionRequest: &common.SubmitTransactionRequest{},
	}

	resp, err := ha.SubmitTransaction(hb.PeerID(), msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
