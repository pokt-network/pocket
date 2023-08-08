package rpc

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/pokt-network/pocket/internal/testutil"
	"github.com/pokt-network/pocket/logger"
	"github.com/pokt-network/pocket/runtime/test_artifacts"
	"github.com/pokt-network/pocket/shared/modules"
	"github.com/stretchr/testify/require"

	"github.com/labstack/echo/v4"
)

func Test_RPCPostV1NodeBackup(t *testing.T) {
	type testCase struct {
		name    string
		setup   func(t *testing.T, e echo.Context) *rpcServer
		assert  func(t *testing.T, tt testCase, e echo.Context, s *rpcServer)
		wantErr bool
	}

	// NB: testDir is used and cleared by each test case
	var testDir = t.TempDir()

	tests := []testCase{
		{
			name: "should create a backup in the specified directory",
			setup: func(t *testing.T, e echo.Context) *rpcServer {
				_, _, url := test_artifacts.SetupPostgresDocker()
				pmod := testutil.NewTestPersistenceModule(t, url)

				s := &rpcServer{
					logger: *logger.Global.CreateLoggerForModule(modules.RPCModuleName),
				}

				s.SetBus(pmod.GetBus())

				e.SetParamNames("dir")
				e.SetParamValues(testDir)

				return s
			},
			wantErr: false,
			assert: func(t *testing.T, tt testCase, e echo.Context, s *rpcServer) {
				f, err := os.Open(testDir)
				require.NoError(t, err)
				dirs, err := f.ReadDir(-1)
				require.NoError(t, err)
				// assert that we wrote the expected 12 files into this directory
				require.True(t, len(dirs) == 12)

				// assert worldstate.json was written
				_, err = os.Open(filepath.Join(testDir, "worldstate.json"))
				require.NoError(t, err)

				// assert blockstore was written
				_, err = os.Open(filepath.Join(testDir, "blockstore.bak"))
				require.NoError(t, err)

				// cleanup the directory after each test
				t.Cleanup(func() {
					require.NoError(t, os.RemoveAll(testDir))
				})
			},
		},
		{
			name: "should error if no directory specified",
			setup: func(t *testing.T, e echo.Context) *rpcServer {
				_, _, url := test_artifacts.SetupPostgresDocker()
				pmod := testutil.NewTestPersistenceModule(t, url)

				s := &rpcServer{
					logger: *logger.Global.CreateLoggerForModule(modules.RPCModuleName),
				}

				s.SetBus(pmod.GetBus())

				return s
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/v1/node/backup", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			s := tt.setup(t, c)

			if err := s.PostV1NodeBackup(c); (err != nil) != tt.wantErr {
				t.Errorf("rpcServer.PostV1NodeBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.assert != nil {
				tt.assert(t, tt, c, s)
			}
		})
	}
}
