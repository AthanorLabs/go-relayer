package contracts

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/AthanorLabs/go-relayer/common"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

// NODE_OPTIONS="--max_old_space_size=8192" ganache --deterministic --accounts=50
var ganachePrivateKeys = []string{
	"4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d",
	"6cbed15c793ce57650b9877cf6fa156fbef513c4e6134f022a85b1ffdd59b2a1",
}

func setupAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	t.Cleanup(func() {
		ec.Close()
	})
	chainID, err := ec.ChainID(context.Background())
	require.NoError(t, err)

	pk, err := ethcrypto.HexToECDSA(ganachePrivateKeys[0])
	require.NoError(t, err)

	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)
	return auth, ec, pk
}

func TestForwarder_Verify(t *testing.T) {
	auth, conn, pk := setupAuth(t)
	chainID, err := conn.ChainID(context.Background())
	require.NoError(t, err)

	address, tx, contract, err := DeployMinimalForwarder(auth, conn)
	require.NoError(t, err)
	require.NotEqual(t, ethcommon.Address{}, address)
	require.NotNil(t, tx)
	require.NotNil(t, contract)
	receipt, err := block.WaitForReceipt(context.Background(), conn, tx.Hash())
	require.NoError(t, err)
	t.Logf("gas cost to deploy MinimalForwarder.sol: %d", receipt.GasUsed)

	key := common.NewKeyFromPrivateKey(pk)

	req := &MinimalForwarderForwardRequest{
		From:  key.Address(),
		To:    ethcommon.Address{2}, // arbitrary
		Value: big.NewInt(0),
		Gas:   big.NewInt(7000000),
		Nonce: big.NewInt(0),
		Data:  []byte{},
	}

	digest, err := GetForwardRequestDigestToSign(req, chainID, address)
	require.NoError(t, err)

	sig, err := key.Sign(digest)
	require.NoError(t, err)

	callOpts := &bind.CallOpts{
		From:    key.Address(),
		Context: context.Background(),
	}

	ok, err := contract.Verify(callOpts, *req, sig)
	require.NoError(t, err)
	require.True(t, ok)
}
