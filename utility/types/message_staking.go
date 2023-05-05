package types

import (
	"net/url"
	"strconv"
	"strings"

	coreTypes "github.com/pokt-network/pocket/shared/core/types"
	"github.com/pokt-network/pocket/shared/utils"
)

// This file captures basic logic common across all the actors that need to stake regardless of their responsibility.

// CLEANUP: Cleanup these strings. Either move them to a shared location or use them in place, but having
// them as constants in this file only feels very incorrect.
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

func validateStaker(msg stakingMessage) coreTypes.Error {
	if err := validateActorType(msg.GetActorType()); err != nil {
		return err
	}
	if err := validateAmount(msg.GetAmount()); err != nil {
		return err
	}
	if err := validateRelayChains(msg.GetChains()); err != nil {
		return err
	}
	return validateServiceURL(msg.GetActorType(), msg.GetServiceUrl())
}

func validateActorType(actorType coreTypes.ActorType) coreTypes.Error {
	if actorType == coreTypes.ActorType_ACTOR_TYPE_UNSPECIFIED {
		return coreTypes.ErrUnknownActorType(string(actorType))
	}
	return nil
}

func validateAmount(amount string) coreTypes.Error {
	if amount == "" {
		return coreTypes.ErrEmptyAmount()
	}
	if _, err := utils.StringToBigInt(amount); err != nil {
		return coreTypes.ErrStringToBigInt(err)
	}
	return nil
}

func validateServiceURL(actorType coreTypes.ActorType, uri string) coreTypes.Error {
	if actorType == coreTypes.ActorType_ACTOR_TYPE_APP {
		return nil
	}

	uri = strings.ToLower(uri)
	_, err := url.ParseRequestURI(uri)
	if err != nil {
		return coreTypes.ErrInvalidServiceURL(err.Error())
	}
	if !(uri[:8] == httpsPrefix || uri[:7] == httpPrefix) {
		return coreTypes.ErrInvalidServiceURL(invalidURLPrefix)
	}

	urlParts := strings.Split(uri, colon)
	if len(urlParts) != 3 { // protocol:host:port
		return coreTypes.ErrInvalidServiceURL(portRequired)
	}
	port, err := strconv.Atoi(urlParts[2])
	if err != nil {
		return coreTypes.ErrInvalidServiceURL(nonNumberPort)
	}
	if port > maxPort || port < 0 {
		return coreTypes.ErrInvalidServiceURL(portOutOfRange)
	}
	if !strings.Contains(uri, period) {
		return coreTypes.ErrInvalidServiceURL(noPeriod)
	}
	return nil
}
