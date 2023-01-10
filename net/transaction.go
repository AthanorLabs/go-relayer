package net

import (
	"context"
	"fmt"
	"time"

	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	"github.com/athanorlabs/go-relayer/common"
)

const (
	transactionID      = "/transaction/0"
	transactionTimeout = time.Second * 5
)

func (h *Host) handleTransactionStream(stream libp2pnetwork.Stream) {
	msg, err := readStreamMessage(stream, maxMessageSize)
	if err != nil {
		log.Debugf("error reading TransactionRequest: %w", err)
	}

	req, ok := msg.(*TransactionRequest)
	if !ok {
		return
	}

	var msgResp *TransactionResponse
	resp, err := h.handleTransaction(req.SubmitTransactionRequest)
	if err != nil {
		msgResp = &TransactionResponse{
			Error: err,
		}
	} else {
		msgResp = &TransactionResponse{
			SubmitTransactionResponse: resp,
		}
	}

	if err := p2pnet.WriteStreamMessage(stream, msgResp, stream.Conn().RemotePeer()); err != nil {
		log.Warnf("failed to send TransactionResponse message to peer: %s", err)
	}

	_ = stream.Close()
}

func (h *Host) SubmitTransaction(who peer.ID, msg *TransactionRequest) (*common.SubmitTransactionResponse, error) {
	ctx, cancel := context.WithTimeout(h.ctx, transactionTimeout)
	defer cancel()

	if err := h.h.Connect(ctx, peer.AddrInfo{ID: who}); err != nil {
		return nil, err
	}

	stream, err := h.h.NewStream(ctx, who, protocol.ID(transactionID))
	if err != nil {
		return nil, fmt.Errorf("failed to open stream with peer: %w", err)
	}

	log.Debugf("opened query stream: %s", stream.Conn())
	defer func() {
		_ = stream.Close()
	}()

	resp, err := submitTransaction(msg, stream)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("received error from remote peer: %w", resp.Error)
	}

	return resp.SubmitTransactionResponse, nil
}

func submitTransaction(msg *TransactionRequest, stream libp2pnetwork.Stream) (*TransactionResponse, error) {
	err := p2pnet.WriteStreamMessage(stream, msg, stream.Conn().RemotePeer())
	if err != nil {
		return nil, err
	}

	return receiveTransactionResponse(stream)
}

func receiveTransactionResponse(stream libp2pnetwork.Stream) (*TransactionResponse, error) {
	msg, err := readStreamMessage(stream, maxMessageSize)
	if err != nil {
		return nil, fmt.Errorf("error reading TransactionResponse: %w", err)
	}

	resp, ok := msg.(*TransactionResponse)
	if !ok {
		return nil, fmt.Errorf("expected message with type %d but received %d",
			TransactionResponseType,
			msg.Type())
	}

	return resp, nil
}
