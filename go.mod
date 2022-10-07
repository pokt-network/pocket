module github.com/pokt-network/pocket

go 1.18

// See the following link for reasoning on why we need the replacement:
// https://discuss.dgraph.io/t/error-mremap-size-mismatch-on-arm64/15333/8
replace github.com/dgraph-io/ristretto v0.1.0 => github.com/46bit/ristretto v0.1.0-with-arm-fix

require (
	github.com/ProtonMail/go-ecvrf v0.0.1
	github.com/golang/mock v1.6.0
	github.com/jackc/pgx/v4 v4.15.0
	github.com/manifoldco/promptui v0.9.0
	github.com/ory/dockertest v3.3.5+incompatible
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/exp v0.0.0-20220301223727-77fc55f9b6c1
	gonum.org/v1/gonum v0.9.3
	google.golang.org/protobuf v1.28.0
)

require (
	github.com/benbjohnson/clock v1.3.0
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/jackc/pgconn v1.11.0
	github.com/jordanorelli/lexnum v0.0.0-20141216151731-460eeb125754
	github.com/quasilyte/go-ruleguard/dsl v0.3.21
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/containerd/continuity v0.3.0 // indirect
	github.com/dgraph-io/ristretto v0.1.0 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/flatbuffers v1.12.1 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/klauspost/compress v1.12.3 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/opencontainers/runc v1.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.12.2
	github.com/sirupsen/logrus v1.8.1 // indirect
	go.opencensus.io v0.22.5 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	gotest.tools v2.2.0+incompatible // indirect
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
)

require (
	filippo.io/edwards25519 v1.0.0-rc.1 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.10.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220405210540-1e041c57c461 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
