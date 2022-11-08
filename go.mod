module github.com/pokt-network/pocket

go 1.18

// See the following link for reasoning on why we need the replacement:
// https://discuss.dgraph.io/t/error-mremap-size-mismatch-on-arm64/15333/8
replace github.com/dgraph-io/ristretto v0.1.0 => github.com/46bit/ristretto v0.1.0-with-arm-fix

require (
	github.com/ProtonMail/go-ecvrf v0.0.1
	github.com/golang/mock v1.6.0
	github.com/jackc/pgx/v4 v4.17.2
	github.com/manifoldco/promptui v0.9.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/stretchr/testify v1.8.0
	golang.org/x/crypto v0.0.0-20221012134737-56aed061732a
	golang.org/x/exp v0.0.0-20221012211006-4de253d81b95
	gonum.org/v1/gonum v0.12.0
	google.golang.org/protobuf v1.28.1
)

require (
	github.com/benbjohnson/clock v1.3.0
	github.com/celestiaorg/smt v0.2.1-0.20220414134126-dba215ccb884
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/getkin/kin-openapi v0.107.0
	github.com/jackc/pgconn v1.13.0
	github.com/jordanorelli/lexnum v0.0.0-20141216151731-460eeb125754
	github.com/labstack/echo/v4 v4.9.1
	github.com/mitchellh/mapstructure v1.5.0
	github.com/quasilyte/go-ruleguard/dsl v0.3.21
	github.com/spf13/cobra v1.6.0
	github.com/spf13/viper v1.13.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/dgraph-io/ristretto v0.1.1 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/flatbuffers v22.9.29+incompatible // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/klauspost/compress v1.15.11 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.13.0
	github.com/sirupsen/logrus v1.9.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/net v0.0.0-20221014081412-f15817d10f9b // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	filippo.io/edwards25519 v1.0.0 // indirect
	github.com/chzyer/readline v1.5.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/invopop/yaml v0.1.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mailru/easyjson v0.0.0-20190626092158-b2ccc519800e // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.4.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.0.0-20221013171732-95e765b1cc43 // indirect
	golang.org/x/text v0.3.8 // indirect
	golang.org/x/tools v0.1.12 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
