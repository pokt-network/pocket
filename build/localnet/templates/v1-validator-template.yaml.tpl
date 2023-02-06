---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: v1-validator${VALIDATOR_NUMBER}
  namespace: default
spec:
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Delete
  selector:
    matchLabels:
      app: v1-validator${VALIDATOR_NUMBER}
  serviceName: v1-validator${VALIDATOR_NUMBER}
  replicas: 1 # we can't really scale validators in stateful sets, because of 1:1 relationship with the private key
  template:
    metadata:
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9000"
      labels:
        app: v1-validator${VALIDATOR_NUMBER}
        v1-purpose: validator
    spec:
      containers:
        - name: pocket-validator
          image: validator-image
          args:
            - pocket
            - -config=/configs/config.json
            - -genesis=/genesis.json
          ports:
            - containerPort: 8080
              name: consensus
            - containerPort: 50832
              name: rpc
          env:
            - name: POCKET_PRIVATE_KEY
              valueFrom:
                secretKeyRef:
                  name: v1-localnet-validators-private-keys
                  key: "${VALIDATOR_NUMBER}"
            - name: POCKET_CONSENSUS_PRIVATE_KEY
              valueFrom:
                secretKeyRef:
                  name: v1-localnet-validators-private-keys
                  key: "${VALIDATOR_NUMBER}"
            - name: POCKET_P2P_PRIVATE_KEY
              valueFrom:
                secretKeyRef:
                  name: v1-localnet-validators-private-keys
                  key: "${VALIDATOR_NUMBER}"
            - name: POSTGRES_USER
              value: "postgres"
            - name: POSTGRES_PASSWORD
              value: LocalNetPassword
            - name: POSTGRES_HOST
              value: dependencies-postgresql
            - name: POSTGRES_PORT
              value: "5432"
            - name: POSTGRES_DB
              value: "postgres"
            - name: POCKET_PERSISTENCE_POSTGRES_URL
              value: "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)"
            - name: POCKET_PERSISTENCE_NODE_SCHEMA
              value: validator${VALIDATOR_NUMBER}
          volumeMounts:
            - name: config-volume
              mountPath: /configs
            - name: genesis-volume
              mountPath: /genesis.json
              subPath: genesis.json
            - name: validator-storage
              mountPath: /validator-storage
      initContainers:
        - name: wait-for-postgres
          image: busybox
          command:
            [
              "sh",
              "-c",
              "until nc -z dependencies-postgresql 5432; do echo waiting for postgres...; sleep 2; done;",
            ]
      volumes:
        - name: config-volume
          configMap:
            name: v1-validator-default-config
        - name: genesis-volume
          configMap:
            name: v1-localnet-genesis
  volumeClaimTemplates:
    - metadata:
        name: validator-storage
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: v1-validator${VALIDATOR_NUMBER}
  namespace: default
  labels:
    app: v1-validator${VALIDATOR_NUMBER}
spec:
  ports:
    - port: 8080
      name: consensus
    - port: 50832
      name: rpc
  selector:
    app: v1-validator${VALIDATOR_NUMBER}
