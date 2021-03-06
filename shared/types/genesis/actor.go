package genesis

// TECHDEBT: There may be an opportunity to refactor this to use generic or be in a shared proto file.
//           `ActorType` is currently defined in `utility/proto/message.proto`, which is separate
//           from where this interface is defined. We need to think/discuss how to consolidate
//           all the types (protos, interfaces, etc...) into a single place.
type Actor interface {
	GetAddress() []byte
	GetPublicKey() []byte
	GetPaused() bool
	GetStatus() int32
	GetChains() []string
	GetGenericParam() string
	GetStakedTokens() string
	GetPausedHeight() int64
	GetUnstakingHeight() int64
	GetOutput() []byte
}

var _ Actor = &App{}
var _ Actor = &Validator{}
var _ Actor = &ServiceNode{}
var _ Actor = &Fisherman{}

func (v *Validator) GetChains() []string     { return nil }
func (v *Validator) GetGenericParam() string { return v.GetServiceUrl() }

func (a *App) GetGenericParam() string { return a.GetMaxRelays() }

func (sn *ServiceNode) GetGenericParam() string { return sn.GetServiceUrl() }

func (f *Fisherman) GetGenericParam() string { return f.GetServiceUrl() }
