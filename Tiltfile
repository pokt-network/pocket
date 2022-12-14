load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://namespace', 'namespace_create')

# Verify k8s version due to https://github.com/zalando/postgres-operator/issues/2098
k8s_version = decode_json(str(local('kubectl version --output=json')).strip())
if k8s_version['serverVersion']['major'] != '1' or k8s_version['serverVersion']['minor'] >= '24':
  fail('Please downgrade your kubernetes version to 1.23. 1.24+ versions are not supported yet (current blocker: https://github.com/zalando/postgres-operator/issues/2098).')


# List of directories Tilt watches to trigger a hot-reload on changes
deps = [
    'app',
    # 'build',
    'consensus',
    'p2p',
    'persistance',
    'rpc',
    'runtime',
    'shared',
    'telemetry',
    'utility',
    'vendor',
    'logger'
]

# Verify pocket operator is available in the parent directory. We use pocket operator to maintain the workloads.
if not os.path.exists('../pocket-operator'):
  print('../pocket-operator directory not found, cloning the repo..')
  local('git clone git@github.com:pokt-network/pocket-operator.git ../pocket-operator')

# TODO(@okdas): add check if the pocket-operator directory has no changes vs the remote and is behind the remote
# to pull the latest changes. This will allow to iterate on operator at the same time as having working localnet,
# but will also allow to get latest changes from the operator repo for developers who don't work on operator.
# Possbly auto check out a newer git tag if already on older one.
include('../pocket-operator/Tiltfile')

# Validators require postgres database - let's install the operator that manages databases in cluster.
is_psql_operator_installed = str(local('kubectl api-resources | grep postgres | grep zalan.do | wc -l')).strip()
if is_psql_operator_installed == '0':
  print('Installing postgres operator')
  local('kubectl apply -k github.com/zalando/postgres-operator/manifests')
# Wait for postgres operator to be available before deploying the database.
local_resource('wait-for-postgres-operator', 'kubectl wait --for=condition=available --timeout=600s --namespace=default deployment postgres-operator')
k8s_yaml('build/localnet/postgres-database.yaml')

# Deploy observability stack (grafana, prometheus, loki) and wire it up with localnet
# TODO(@okdas): check if helm cli is available.
helm_repo('grafana', 'https://grafana.github.io/helm-charts', resource_name='helm-repo-grafana')
helm_repo('prometheus-community', 'https://prometheus-community.github.io/helm-charts', resource_name='helm-repo-prometheus')

if not os.path.exists('build/localnet/observability-stack/charts'):
    local('helm dependency build build/localnet/observability-stack')

namespace_create('observability')
k8s_yaml(helm("build/localnet/observability-stack", name='observability-stack', namespace="observability"))

# Builds the pocket binary. Note target OS is linux, because it later will be run in a container.
local_resource('pocket: Watch & Compile', 'GOOS=linux go build -o bin/pocket-linux app/pocket/main.go', deps=deps)
local_resource('debug client: Watch & Compile', 'GOOS=linux go build -tags=debug -o bin/client-linux app/client/*.go', deps=deps)

# Builds and maintains the validator container image after the binary is built on local machine
docker_build('validator-image', '.',
    dockerfile_contents='''FROM debian:bullseye
COPY build/localnet/start.sh /start.sh
COPY build/localnet/restart.sh /restart.sh
COPY bin/pocket-linux /usr/local/bin/pocket
WORKDIR /
CMD ["/usr/local/bin/pocket"]
ENTRYPOINT ["/start.sh", "/usr/local/bin/pocket", "-config=/configs/config.json", "-genesis=/genesis.json"]
''',
    only=['./bin/pocket-linux', './build/'],
    live_update=[
        sync('./bin/pocket-linux', '/usr/local/bin/pocket'),
        run('/restart.sh'), # TODO(@okdas): add healthchecks as this is possibly catching some of the issues with running the validators, e.g. when postgres db is not provisioned yet?
    ]
)

# Builds and maintains the client container image after the binary is built on local machine
docker_build('client-image', '.',
    dockerfile_contents='''FROM debian:bullseye
WORKDIR /
COPY bin/client-linux /usr/local/bin/client
CMD ["/usr/local/bin/client"]
''',
    only=['./bin/client-linux'],
    live_update=[
        sync('./bin/client-linux', '/usr/local/bin/client'),
    ]
)

# Makes Tilt aware of our own Custom Resource Definition from pocket-operator, so it can work with our operator.
k8s_kind('PocketValidator', image_json_path='{.spec.pocketImage}')

# Wait for postgres database to be available before deploying the validators.
local_resource('wait-for-postgres-database', 'until kubectl wait postgresqls --for=jsonpath={.status.PostgresClusterStatus}=Running pocket-database; do sleep 3; done')

# Wait for pocket operator
local_resource('wait-for-pocket-operator', 'kubectl wait --for=condition=available --timeout=600s --namespace=pocket-operator-system deployment pocket-operator-controller-manager')

# TODO(@okdas): https://github.com/tilt-dev/tilt/issues/3048
# Pushes localnet manifests to the cluster.
k8s_yaml([
    'build/localnet/private-keys.yaml',
    'build/localnet/validators.yaml',
    'build/localnet/cli-client.yaml',
    'build/localnet/network.yaml'])

# Exposes postgres port to 5432 on the host machine.
k8s_resource(new_name='postgres',
             objects=['pocket-database:postgresql'],
             extra_pod_selectors=[{'cluster-name': 'pocket-database'}],
             port_forwards=5432)

# Exposes grafana
k8s_resource(new_name='grafana',
             workload='observability-stack-grafana',
             extra_pod_selectors=[{'app.kubernetes.io/name': 'grafana'}],
             port_forwards=['42000:3000'])
