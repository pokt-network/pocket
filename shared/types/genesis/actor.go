package genesis

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

	GetPoolName() string
}

var _ Actor = &App{}
var _ Actor = &Validator{}
var _ Actor = &ServiceNode{}
var _ Actor = &Fisherman{}

func (v *Validator) GetChains() []string     { return nil }
func (v *Validator) GetGenericParam() string { return v.GetServiceUrl() }
func (v *Validator) GetPoolName() string     { return ValidatorStakePoolName }

func (a *App) GetGenericParam() string { return a.GetMaxRelays() }
func (a *App) GetPoolName() string     { return AppStakePoolName }

func (sn *ServiceNode) GetGenericParam() string { return sn.GetServiceUrl() }
func (sn *ServiceNode) GetPoolName() string     { return ServiceNodeStakePoolName }

func (f *Fisherman) GetGenericParam() string { return f.GetServiceUrl() }
func (f *Fisherman) GetPoolName() string     { return FishermanStakePoolName }
