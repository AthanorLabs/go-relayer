package net

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/athanorlabs/go-relayer/common"
)

// Message must be implemented by all network messages
type Message interface {
	Type() byte
}

func decodeMessage(b []byte) (Message, error) {
	// 1-byte type followed by at least 2-bytes of JSON (`{}`)
	if len(b) < 3 {
		return nil, errors.New("invalid message bytes")
	}

	msgType := b[0]
	msgJSON := b[1:]
	var msg Message

	switch msgType {
	case TransactionResponseType:
		msg = &TransactionResponse{}
	case TransactionRequestType:
		msg = &TransactionRequest{}
	default:
		return nil, fmt.Errorf("invalid message type %d", msgType)
	}

	if err := json.Unmarshal(msgJSON, &msg); err != nil {
		return nil, fmt.Errorf("failed to decode message with type %d: %w", msg.Type(), err)
	}
	return msg, nil
}

const (
	TransactionRequestType byte = iota
	TransactionResponseType
)

type TransactionRequest struct {
	*common.SubmitTransactionRequest
}

// String ...
func (m *TransactionRequest) String() string {
	return fmt.Sprintf("TransactionRequest %v", m.SubmitTransactionRequest)
}

// Encode ...
func (m *TransactionRequest) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{TransactionRequestType}, b...), nil
}

func (r *TransactionRequest) Type() byte {
	return TransactionRequestType
}

type TransactionResponse struct {
	*common.SubmitTransactionResponse
	Error error
}

// String ...
func (m *TransactionResponse) String() string {
	if m.Error != nil {
		return fmt.Sprintf("TransactionResponse %s", m.Error)
	}

	return fmt.Sprintf("TransactionResponse %v", m.SubmitTransactionResponse)
}

// Encode ...
func (m *TransactionResponse) Encode() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{TransactionResponseType}, b...), nil
}

func (r *TransactionResponse) Type() byte {
	return TransactionResponseType
}
