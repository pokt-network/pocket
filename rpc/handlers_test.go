package rpc

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/modules"
)

func Test_RPCPostV1NodeBackup(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, e echo.Context) *rpcServer
		wantErr bool
	}{
		{
			name: "should create a backup in the default directory",
			setup: func(t *testing.T, e echo.Context) *rpcServer {
				_, _, url := test_artifacts.SetupPostgresDocker()
				pmod := testutil.NewTestPersistenceModule(t, url)
				// context := testutil.NewTestPostgresContext(t, pmod, 0)

				s := &rpcServer{
					logger: *logger.Global.CreateLoggerForModule(modules.RPCModuleName),
				}

				s.SetBus(pmod.GetBus())

				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new echo Context for each test
			tempDir := t.TempDir()
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, tempDir, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// pass the fresh context to setup
			s := tt.setup(t, c)

			// call and assert
			if err := s.PostV1NodeBackup(c); (err != nil) != tt.wantErr {
				t.Errorf("rpcServer.PostV1NodeBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
