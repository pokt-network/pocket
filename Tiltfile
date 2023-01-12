load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://namespace', 'namespace_create')
load('ext://restart_process', 'docker_build_with_restart')

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

# Deploy dependencies (grafana, postgres, prometheus) and wire it up with localnet
helm_repo('grafana', 'https://grafana.github.io/helm-charts', resource_name='helm-repo-grafana')
helm_repo('prometheus-community', 'https://prometheus-community.github.io/helm-charts', resource_name='helm-repo-prometheus')
helm_repo('bitnami', 'https://charts.bitnami.com/bitnami', resource_name='helm-repo-bitnami')

# Avoid downloading dependencies if no missing/outdated charts
check_helm_dependencies = local("helm dependency list build/localnet/dependencies | awk '{print $4}' | grep -Ev 'ok|STATUS'")
helm_dependencies_not_ok_count = len(str(check_helm_dependencies).splitlines())
if helm_dependencies_not_ok_count > 1:
    local('helm dependency update build/localnet/dependencies')

k8s_yaml(helm("build/localnet/dependencies", name='dependencies', namespace="default"))

# Builds the pocket binary. Note target OS is linux, because no matter what your OS is, container runs linux natively or in VM.
local_resource('pocket: Watch & Compile', 'GOOS=linux go build -o bin/pocket-linux app/pocket/main.go', deps=deps)
local_resource('debug client: Watch & Compile', 'GOOS=linux go build -tags=debug -o bin/client-linux app/client/*.go', deps=deps)

# Builds and maintains the validator container image after the binary is built on local machine, restarts a process on code change
docker_build_with_restart('validator-image', '.',
    dockerfile_contents='''FROM debian:bullseye
COPY bin/pocket-linux /usr/local/bin/pocket
WORKDIR /
''',
    only=['./bin/pocket-linux'],
    entrypoint=["/usr/local/bin/pocket", "-config=/configs/config.json", "-genesis=/genesis.json"],
    live_update=[
        sync('bin/pocket-linux', '/usr/local/bin/pocket')
    ]
)

# Builds and maintains the client container image after the binary is built on local machine
docker_build_with_restart('client-image', '.',
    dockerfile_contents='''FROM debian:bullseye
WORKDIR /
COPY bin/client-linux /usr/local/bin/client
''',
    only=['bin/client-linux'],
    entrypoint=["sleep", "infinity"],
    live_update=[
        sync('bin/client-linux', '/usr/local/bin/client')
    ]
)

# TODO(@okdas): https://github.com/tilt-dev/tilt/issues/3048
# Pushes localnet manifests to the cluster.
k8s_yaml([
    'build/localnet/private-keys.yaml',
    'build/localnet/v1-validator1.yaml',
    'build/localnet/v1-validator2.yaml',
    'build/localnet/v1-validator3.yaml',
    'build/localnet/v1-validator4.yaml',
    'build/localnet/configs.yaml',
    'build/localnet/cli-client.yaml',
    'build/localnet/network.yaml'])

# # Exposes postgres port to 5432 on the host machine.
# k8s_resource(new_name='postgres',
#              objects=['pocket-database:postgresql'],
#              extra_pod_selectors=[{'cluster-name': 'pocket-database'}],
#              port_forwards=5432)

# Exposes grafana
k8s_resource(new_name='grafana',
             workload='dependencies-grafana',
             extra_pod_selectors=[{'app.kubernetes.io/name': 'grafana'}],
             port_forwards=['42000:3000'])
