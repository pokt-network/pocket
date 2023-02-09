package types

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/pokt-network/pocket/shared/converters"
	coreTypes "github.com/pokt-network/pocket/shared/core/types"
)

const (
	httpsPrefix      = "https://"
	httpPrefix       = "http://"
	colon            = ":"
	period           = "."
	invalidURLPrefix = "the url must start with http:// or https://"
	portRequired     = "a port is required"
	nonNumberPort    = "invalid port, cant convert to integer"
	portOutOfRange   = "invalid port, out of valid port range"
	noPeriod         = "must contain one '.'"
	maxPort          = 65535
)

// This interface is useful in validating stake related messages and is not intended to be used outside of this package
type stakingMessage interface {
	GetActorType() coreTypes.ActorType
	GetAmount() string
	GetChains() []string
	GetServiceUrl() string
}

func validateStaker(msg stakingMessage) Error {
	if err := validateActorType(msg.GetActorType()); err != nil {
		return err
	}
	if err := validateAmount(msg.GetAmount()); err != nil {
		return err
	}
	if err := validateRelayChains(msg.GetChains()); err != nil {
		return err
	}
	return validateServiceUrl(msg.GetActorType(), msg.GetServiceUrl())
}

func validateActorType(actorType coreTypes.ActorType) Error {
	if actorType == coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED {
		return ErrUnknownActorType(string(actorType))
	}
	return nil
}

func validateAmount(amount string) Error {
	if amount == "" {
		return ErrEmptyAmount()
	}
	if _, err := converters.StringToBigInt(amount); err != nil {
		return ErrStringToBigInt(err)
	}
	return nil
}

func validateServiceUrl(actorType coreTypes.ActorType, uri string) Error {
	if actorType == coreTypes.ActorType_ACTOR_TYPE_APP {
		return nil
	}

	uri = strings.ToLower(uri)
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return ErrInvalidServiceUrl(err.Error())
	}
	if !(uri[:8] == httpsPrefix || uri[:7] == httpPrefix) {
		return ErrInvalidServiceUrl(invalidURLPrefix)
	}

	urlParts := strings.Split(uri, colon)
	if len(urlParts) != 3 { // protocol:host:port
		return ErrInvalidServiceUrl(portRequired)
	}
	port, err := strconv.Atoi(urlParts[2])
	if err != nil {
		return ErrInvalidServiceUrl(nonNumberPort)
	}
	if port > maxPort || port < 0 {
		return ErrInvalidServiceUrl(portOutOfRange)
	}
	if !strings.Contains(uri, period) {
		return ErrInvalidServiceUrl(noPeriod)
	}
	return nil
}
