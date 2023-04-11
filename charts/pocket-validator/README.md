<div align="center">
  <a href="https://www.pokt.network">
    <img src="https://user-images.githubusercontent.com/2219004/151564884-212c0e40-3bfa-412e-a341-edb54b5f1498.jpeg" alt="Pocket Network logo" width="340"/>
  </a>
</div>

# pocket-validator

Validator for Pocket Network - decentralized blockchain infrastructure

## Requirements

### Private key

In order to use this chart, you must have a Pocket Network wallet with a private key (make sure you made a backup!).
If you do not have a private key, you can create one by following the instructions [here](https://docs.pokt.network/pokt/wallets/#create-wallet).
This helm chart assumes user utilizes Kubernetes Secret to store the private key for an additional layer of protection. The key should not be protected with password.

Here is an example of the private key stored in a Kubernetes Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
    name: validator-private-key
stringData:
    "1919605e50c0a60177d0554b528c9810313523b3": "4d6d24690137b0c43dee3490cafa4ca49cc1c4facdd1a73be1255a5b752223dc2b7672ea2493dcdd0efc6c6caf1073c4f3ff8508c686031e2d1244c02f0b900d"
```

This secret can then be utilized with this helm chart using the following variables:

```yaml
privateKeySecretKeyRef:
  name: validator-private-key
  key: 1919605e50c0a60177d0554b528c9810313523b3
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| config.consensus.max_mempool_bytes | int | `500000000` |  |
| config.consensus.pacemaker_config.debug_time_between_steps_msec | int | `1000` |  |
| config.consensus.pacemaker_config.manual | bool | `true` |  |
| config.consensus.pacemaker_config.timeout_msec | int | `10000` |  |
| config.consensus.private_key | string | `""` |  |
| config.logger.format | string | `"json"` |  |
| config.logger.level | string | `"debug"` |  |
| config.p2p.is_empty_connection_type | bool | `false` |  |
| config.p2p.max_mempool_count | int | `100000` |  |
| config.p2p.port | int | `42069` |  |
| config.p2p.private_key | string | `""` |  |
| config.p2p.use_rain_tree | bool | `true` |  |
| config.persistence.block_store_path | string | `"/pocket/validator/block-store"` |  |
| config.persistence.health_check_period | string | `"30s"` |  |
| config.persistence.max_conn_idle_time | string | `"1m"` |  |
| config.persistence.max_conn_lifetime | string | `"5m"` |  |
| config.persistence.max_conns_count | int | `50` |  |
| config.persistence.min_conns_count | int | `1` |  |
| config.persistence.node_schema | string | `"validator"` |  |
| config.persistence.postgres_url | string | `""` |  |
| config.persistence.trees_store_dir | string | `"/pocket/validator/trees"` |  |
| config.persistence.tx_indexer_path | string | `"/pocket/validator/tx-indexer"` |  |
| config.private_key | string | `""` |  |
| config.root_directory | string | `"/go/src/github.com/pocket-network"` |  |
| config.rpc.enabled | bool | `true` |  |
| config.rpc.port | string | `"50832"` |  |
| config.rpc.timeout | int | `30000` |  |
| config.rpc.use_cors | bool | `false` |  |
| config.telemetry.address | string | `"0.0.0.0:9000"` |  |
| config.telemetry.enabled | bool | `true` |  |
| config.telemetry.endpoint | string | `"/metrics"` |  |
| config.use_libp2p | bool | `false` |  |
| config.utility.max_mempool_transaction_bytes | int | `1073741824` |  |
| config.utility.max_mempool_transactions | int | `9000` |  |
| externalPostgresql.database | string | `""` | name of the external database |
| externalPostgresql.enabled | bool | `false` | use external postgres database |
| externalPostgresql.host | string | `""` | host of the external database |
| externalPostgresql.passwordSecretKeyRef.key | string | `""` | key in the Secret that contains the database password |
| externalPostgresql.passwordSecretKeyRef.name | string | `""` | name of the Secret in the same namespace that contains the database password |
| externalPostgresql.port | int | `5432` | port of the external database |
| externalPostgresql.userSecretKeyRef.key | string | `""` | key in the Secret that contains the database user |
| externalPostgresql.userSecretKeyRef.name | string | `""` | name of the Secret in the same namespace that contains the database user |
| fullnameOverride | string | `""` |  |
| genesis.externalConfigMap.key | string | `""` | Key in the ConfigMap that contains the genesis file, only used if `genesis.preProvisionedGenesis.enabled` is false |
| genesis.externalConfigMap.name | string | `""` | Name of the ConfigMap that contains the genesis file, only used if `genesis.preProvisionedGenesis.enabled` is false |
| genesis.preProvisionedGenesis.enabled | bool | `true` | Use genesis file supplied by the Helm chart, of false refer to `genesis.externalConfigMap` |
| genesis.preProvisionedGenesis.type | string | `"devnet"` | Type of the genesis file to use, can be `devnet`, `testnet`, `mainnet` |
| global.postgresql.service.ports.postgresql | string | `"5432"` |  |
| image.pullPolicy | string | `"IfNotPresent"` | image pull policy |
| image.repository | string | `"ghcr.io/pokt-network/pocket-v1"` | image repository |
| image.tag | string | `"latest"` | image tag |
| imagePullSecrets | list | `[]` | image pull secrets |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` | enable ingress for RPC port |
| ingress.hosts[0].host | string | `"chart-example.local"` |  |
| ingress.hosts[0].paths[0].path | string | `"/"` |  |
| ingress.hosts[0].paths[0].pathType | string | `"ImplementationSpecific"` |  |
| ingress.tls | list | `[]` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| persistence.accessModes | list | `["ReadWriteOnce"]` | persistent Volume Access Modes |
| persistence.annotations | object | `{}` | annotations of the persistent volume claim |
| persistence.dataSource | object | `{}` | custom data source of the persistent volume claim |
| persistence.enabled | bool | `true` | enable persistent volume claim |
| persistence.existingClaim | string | `""` | name of an existing PVC to use for persistence |
| persistence.reclaimPolicy | string | `"Delete"` | persistent volume reclaim policy |
| persistence.selector | object | `{}` | selector to match an existing Persistent Volume |
| persistence.size | string | `"8Gi"` | size of the persistent volume claim |
| persistence.storageClass | string | `""` | storage class of the persistent volume claim |
| podAnnotations | object | `{}` | pod annotations |
| podSecurityContext | object | `{}` |  |
| postgresql.enabled | bool | `true` | deploy postgresql database automatically. Refer to https://github.com/bitnami/charts/blob/main/bitnami/postgresql/values.yaml for additional options. |
| postgresql.primary.persistence.enabled | bool | `false` | enable persistent volume claim for PostgreSQL |
| postgresql.primary.persistence.size | string | `"8Gi"` | size of the persistent volume claim for PostgreSQL |
| privateKeySecretKeyRef.key | string | `""` | REQUIRED. Key in the Secret that contains the private key of the node |
| privateKeySecretKeyRef.name | string | `""` | REQUIRED. Name of the Secret in the same namespace that contains the private key of the node |
| resources | object | `{}` | resources limits and requests |
| securityContext | object | `{}` |  |
| service.annotations | object | `{}` | service annotations |
| service.ports.consensus | int | `42069` | consensus port of the node |
| service.ports.metrics | int | `9000` | OpenTelemetry metrics port of the node |
| service.ports.rpc | int | `50832` | rpc port of the node |
| service.type | string | `"ClusterIP"` | service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |
| tolerations | list | `[]` |  |
