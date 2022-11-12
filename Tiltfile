# TODO: add resource dependencies

# Verify pocket operator is available in the parent directory. We use pocket operator to maintain the workloads.
if not os.path.exists('../pocket-operator'):
  fail('Please "git clone" the git@github.com:pokt-network/pocket-operator.git repo in ../pocket-operator!')

# TODO(@okdas): add check if the pocket-operator directory has no changes vs the remote and is behind the remote
# to pull the latest changes. This will allow to iterate on operator at the same time as having working localnet,
# but will also allow to get latest changes from the operator repo for developers who don't work on operator.
include('../pocket-operator/Tiltfile')

# Validators require postgres database - let's install the operator that manages databases in cluster.
is_psql_operator_installed = str(local('kubectl api-resources | grep postgres | grep zalan.do | wc -l')).strip()
if is_psql_operator_installed == '0':
  print('Installing postgres operator')
  local('kubectl apply -k github.com/zalando/postgres-operator/manifests')

# Builds the pocket binary. Note target OS is linux, because it later will be run in a container.
local_resource('pocket: Watch & Compile', 'GOOS=linux go build -o bin/pocket-linux app/pocket/main.go', deps=['app/pocket/main.go'])
local_resource('client: Watch & Compile', 'GOOS=linux go build -o bin/client-linux app/client/main.go', deps=['**/**.go'])

# Builds and maintains the container after the binary is built on machine
docker_build('validator-image', '.',
    dockerfile_contents='''FROM debian:bullseye
COPY build/localnet/start.sh /start.sh
COPY build/localnet/restart.sh /restart.sh
COPY bin/pocket-linux /usr/local/bin/pocket
WORKDIR /
CMD ["/usr/local/bin/pocket"]
ENTRYPOINT ["/start.sh", "/usr/local/bin/pocket"]
''',
    only=['./bin/pocket-linux'],
    live_update=[
        sync('./bin/pocket-linux', '/usr/local/bin/pocket'),
        run('/restart.sh'),
    ]
)

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

# Makes Tilt aware of our own custom resource, so it can work with our operator.
k8s_kind('PocketValidator', image_json_path='{.spec.pocketImage}')

# Pushes localnet manifests to the cluster.
k8s_yaml([
    'build/localnet/postgres-database.yaml',
    'build/localnet/private-keys.yaml',
    'build/localnet/validators.yaml',
    'build/localnet/cli-client.yaml',
    'build/localnet/network.yaml'])

k8s_resource(new_name='postgres',
             objects=['pocket-database:postgresql'],
             extra_pod_selectors=[{'cluster-name': 'pocket-database'}],
             port_forwards=5432)

