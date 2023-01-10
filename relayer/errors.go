package relayer

import (
	"errors"
)

var (
	errMustSetForwarder               = errors.New("must set Forwarder")
	errMustSetNewForwardRequestFunc   = errors.New("must set NewForwardRequestFunc")
	errMustSetValidateTransactionFunc = errors.New("must set ValidateTransactionFunc")
	errMustSetKey                     = errors.New("must set Key")
	errMustSetEthClient               = errors.New("must set EthClient")

	errFailedToVerify       = errors.New("failed to verify forward request signature")
	errFromNotProvided      = errors.New("must provide `from` field in request")
	errSignatureNotProvided = errors.New("must provide `signature` field in request")
	errGasNotProvided       = errors.New("must provide `gas` field in request")
	errNonceNotProvided     = errors.New("must provide `nonce` field in request")
)
