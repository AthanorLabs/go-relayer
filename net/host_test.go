package net

import (
	"context"
	"path"
	"testing"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	"github.com/athanorlabs/go-relayer/common"

	"github.com/stretchr/testify/require"
)

func testHandleTransactionFunc(*common.SubmitTransactionRequest) (*common.SubmitTransactionResponse, error) {
	return &common.SubmitTransactionResponse{}, nil
}

func basicTestConfig(t *testing.T) *Config {
	// t.TempDir() is unique on every call. Don't reuse this config with multiple hosts.
	tmpDir := t.TempDir()
	netCfg := &p2pnet.Config{
		Ctx:        context.Background(),
		DataDir:    tmpDir,
		Port:       0, // OS randomized libp2p port
		KeyFile:    path.Join(tmpDir, "node.key"),
		Bootnodes:  nil,
		ProtocolID: "/testid",
		ListenIP:   "127.0.0.1",
	}

	return &Config{
		P2pConfig:             netCfg,
		HandleTransactionFunc: testHandleTransactionFunc,
		IsRelayer:             true,
	}
}

func newHost(t *testing.T, cfg *Config) *Host {
	h, err := NewHost(cfg)
	require.NoError(t, err)

	err = h.Start()
	require.NoError(t, err)

	t.Cleanup(func() {
		err = h.Stop()
		require.NoError(t, err)
	})
	return h
}

func TestNewHost(t *testing.T) {
	_ = newHost(t, basicTestConfig(t))
}
