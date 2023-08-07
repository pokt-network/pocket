package rpc

import (
	"io"
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
	// THIS WORKS BUT ADJUST LATER
	type testCase struct {
		name    string
		setup   func(t *testing.T, e echo.Context) *rpcServer
		assert  func(t *testing.T, tt testCase, e echo.Context, s *rpcServer)
		wantErr bool
	}

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
				empty, err := isEmpty(testDir)
				require.NoError(t, err)
				require.False(t, empty)
				f, err := os.Open(testDir)
				require.NoError(t, err)
				dirs, err := f.ReadDir(-1)
				require.NoError(t, err)
				require.True(t, len(dirs) == 12)

				// assert worldstate json was written
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create a new echo Context for each test
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/v1/node/backup", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// pass the fresh context to setup
			s := tt.setup(t, c)

			// call and assert
			if err := s.PostV1NodeBackup(c); (err != nil) != tt.wantErr {
				t.Errorf("rpcServer.PostV1NodeBackup() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.assert(t, tt, c, s)
		})
	}
}

// TECHDEBT(#796) - Organize and dedupe this function into testutil package
func isEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
