package savepoints

import (
	"sync"

	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/shared/modules"
)

const (
	moduleName = "savepoint_factory"

	appsKey       = "apps"
	validatorsKey = "validators"
	fishermenKey  = "fishermen"
	servicersKey  = "servicers"
	accountsKey   = "accounts"
	poolsKey      = "pools"
	paramsKey     = "params"
	flagsKey      = "flags"
)

var _ modules.SavepointFactory = &savepointFactory{}

type savepointFactory struct {
	readContext modules.PersistenceReadContext
	logger      *modules.Logger
}

func NewSavepointFactory(readContext modules.PersistenceReadContext) modules.SavepointFactory {
	return &savepointFactory{
		readContext: readContext,
		logger:      logger.Global.CreateLoggerForModule(moduleName),
	}
}

func (sm *savepointFactory) CreateSavepoint(height int64) (modules.PersistenceReadContext, error) {
	getters := []struct {
		name       string
		getterFunc func(height int64) (string, error)
	}{
		{appsKey, sm.readContext.GetAllAppsJSON},
		{validatorsKey, sm.readContext.GetAllValidatorsJSON},
		{fishermenKey, sm.readContext.GetAllFishermenJSON},
		{servicersKey, sm.readContext.GetAllServicersJSON},
		{accountsKey, sm.readContext.GetAllAccountsJSON},
		{poolsKey, sm.readContext.GetAllPoolsJSON},
		{paramsKey, sm.readContext.GetAllParamsJSON},
		{flagsKey, sm.readContext.GetAllFlagsJSON},
	}

	var mutex sync.Mutex
	var wg sync.WaitGroup
	wg.Add(len(getters))
	errChan := make(chan error, 1)
	cancelChan := make(chan struct{})

	dataDumps := make(map[string]string, len(getters))

	for _, getter := range getters {
		// INVESTIGATE: the code below is structured so that it could run in parallel just by adding `go` to start goroutines below but for some reason the pg.tx is nil
		// when the goroutines are started. This is a temporary fix to get the code working but it should be investigated further.
		func(getter struct {
			name       string
			getterFunc func(height int64) (string, error)
		}) {
			defer wg.Done()

			select {
			case <-cancelChan:
				// exit the goroutine if cancellation is signaled
				return
			default:
				// continue execution if no cancellation signal is received
			}

			data, err := getter.getterFunc(height)
			if err != nil {
				select {
				case errChan <- err:
					sm.logger.Error().Err(err).Str("getter_name", getter.name).Msg("error getting data for savepoint")
					// signal cancellation to all goroutines
					close(cancelChan)
				default:
				}
				return
			}

			mutex.Lock()
			dataDumps[getter.name] = data
			mutex.Unlock()
		}(getter)
	}

	return &savepoint{
		height: height,

		appsJson:       dataDumps[appsKey],
		validatorsJson: dataDumps[validatorsKey],
		fishermenJson:  dataDumps[fishermenKey],
		servicersJson:  dataDumps[servicersKey],
		accountsJson:   dataDumps[accountsKey],
		poolsJson:      dataDumps[poolsKey],
		paramsJson:     dataDumps[paramsKey],
		flagsJson:      dataDumps[flagsKey],
	}, nil
}
