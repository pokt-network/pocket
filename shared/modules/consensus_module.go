package modules

import (
	"google.golang.org/protobuf/types/known/anypb"
)

type ConsensusModule interface {
	Module
	HandleMessage(*anypb.Any)
}
