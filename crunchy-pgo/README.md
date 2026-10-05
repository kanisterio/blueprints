# Crunchy Postgres for Kubernetes (PGO)

[PGO](https://github.com/CrunchyData/postgres-operator), the Postgres Operator from [Crunchy Data](https://www.crunchydata.com), gives you a declarative way to run production-ready PostgreSQL clusters on Kubernetes. It handles high availability, upgrades, monitoring, and backup/restore. Backup and restore are built on [pgBackRest](https://pgbackrest.org).

## Introduction

This blueprint uses the operator's own backup and restore mechanism. Kanister doesn't copy any data itself:

- **Backup**: the blueprint triggers a one-off pgBackRest backup by annotating the `PostgresCluster` resource with `postgres-operator.crunchydata.com/pgbackrest-backup`. It then waits for the backup to succeed and records the backup repo, the current database timestamp and the current timeline ID as Kanister output artifacts.
- **Restore**: the blueprint patches the `PostgresCluster` with a point-in-time restore spec (`--type=time`) that uses the recorded timestamp and timeline. It triggers the restore with the `postgres-operator.crunchydata.com/pgbackrest-restore` annotation, waits for the restore to succeed, and then disables the restore spec again.

Backup data is stored in the pgBackRest repository configured on the `PostgresCluster` (a PVC, S3, GCS or Azure Blob), not in a Kanister Profile location.

## Prerequisites

- Kubernetes 1.25+
- PV provisioner support in the underlying infrastructure
- PGO v5.x installed in your cluster
- Kanister controller version 0.118.0 installed in your cluster, let's assume in Namespace `kanister`
- Kanctl CLI installed (https://docs.kanister.io/tooling.html#install-the-tools)

## Installing PGO

To install the PGO operator using the official Helm chart:

```bash
$ helm install pgo oci://registry.developers.crunchydata.com/crunchydata/pgo \
    --namespace postgres-operator --create-namespace

# Verify the operator is running
$ kubectl get pods -n postgres-operator
NAME                   READY   STATUS    RESTARTS   AGE
pgo-6b9c8f8c79-7xk2p   1/1     Running   0          1m
```

You can find other installation methods (Kustomize, OperatorHub) in the [PGO documentation](https://access.crunchydata.com/documentation/postgres-operator/latest/installation).

## Creating a PostgresCluster

**NOTE:**

The blueprint requires **manual backups** to be enabled on the `PostgresCluster` (`spec.backups.pgbackrest.manual`). The backup action fails if this field isn't set.

Create a `PostgresCluster` named `hippo` in the `pgo-test` namespace with a PVC-based pgBackRest repository and manual backups enabled:

```bash
$ kubectl create namespace pgo-test

$ cat <<EOF | kubectl apply -f -
apiVersion: postgres-operator.crunchydata.com/v1beta1
kind: PostgresCluster
metadata:
  name: hippo
  namespace: pgo-test
spec:
  postgresVersion: 16
  instances:
    - name: instance1
      replicas: 1
      dataVolumeClaimSpec:
        accessModes:
          - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
  backups:
    pgbackrest:
      manual:
        repoName: repo1
        options:
          - --type=full
      repos:
        - name: repo1
          volume:
            volumeClaimSpec:
              accessModes:
                - "ReadWriteOnce"
              resources:
                requests:
                  storage: 1Gi
EOF
```

To store backups in object storage (S3, GCS, Azure) instead of a PVC, configure `spec.backups.pgbackrest.repos` as described in the [PGO backup documentation](https://access.crunchydata.com/documentation/postgres-operator/latest/tutorials/backups-disaster-recovery/backups).

Wait until the cluster and its initial backup are ready:

```bash
$ kubectl get pods -n pgo-test
NAME                      READY   STATUS      RESTARTS   AGE
hippo-backup-xxxx-xxxxx   0/1     Completed   0          2m
hippo-instance1-abcd-0    4/4     Running     0          3m
hippo-repo-host-0         2/2     Running     0          3m
```

## Integrating with Kanister

If you created the `PostgresCluster` with a name other than `hippo` or in a namespace other than `pgo-test`, update the commands below (backup, restore and verification) to use the correct name and namespace.

### Grant Kanister access to PGO resources

The blueprint's `KubeTask` and `Wait` phases read and patch `postgresclusters`, list `statefulsets` and `jobs`, and exec into the database pod. Grant the Kanister controller's service account these permissions:

```bash
$ cat <<EOF | kubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kanister-operator-cluster-role-pgo
  labels:
    app: kanister-operator
rules:
- apiGroups: ["postgres-operator.crunchydata.com"]
  resources: ["postgresclusters"]
  verbs: ["get", "list", "watch", "patch", "update"]
- apiGroups: ["apps"]
  resources: ["statefulsets"]
  verbs: ["get", "list"]
- apiGroups: ["batch"]
  resources: ["jobs"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["pods/exec"]
  verbs: ["create"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kanister-operator-role-pgo
  labels:
    app: kanister-operator
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kanister-operator-cluster-role-pgo
subjects:
- kind: ServiceAccount
  name: kanister-kanister-operator
  namespace: kanister
EOF
```

You may have to update the service account name and namespace in the binding if you deployed Kanister with a different release name or namespace.

### Create Blueprint

Create Blueprint in the same namespace as the Kanister controller

```bash
$ kubectl create -f ./pgo-blueprint.yaml -n kanister
blueprint.cr.kanister.io/pgo-blueprint created
```

Once the PostgresCluster is running, you can populate it with some data. Let's add a table called "company" to a test database:

```bash
# Connect to PostgreSQL by running a shell inside the primary database pod
$ kubectl exec -ti -n pgo-test -c database \
    $(kubectl get pods -n pgo-test --selector=postgres-operator.crunchydata.com/cluster=hippo,postgres-operator.crunchydata.com/role=master -o=jsonpath='{.items[0].metadata.name}') -- bash

# From inside the shell, use psql to insert some data
$ psql

postgres=# CREATE DATABASE test;
CREATE DATABASE

postgres=# \c test
You are now connected to database "test" as user "postgres".

# Create "company" table
test=# CREATE TABLE COMPANY(
     ID INT PRIMARY KEY     NOT NULL,
     NAME           TEXT    NOT NULL,
     AGE            INT     NOT NULL,
     ADDRESS        CHAR(50),
     SALARY         REAL,
     CREATED_AT    TIMESTAMP
);
CREATE TABLE

# Insert data
test=# INSERT INTO COMPANY (ID,NAME,AGE,ADDRESS,SALARY,CREATED_AT) VALUES (10, 'Paul', 32, 'California', 20000.00, now());
INSERT 0 1
test=# INSERT INTO COMPANY (ID,NAME,AGE,ADDRESS,SALARY,CREATED_AT) VALUES (20, 'Omkar', 32, 'California', 20000.00, now());
INSERT 0 1
test=# INSERT INTO COMPANY (ID,NAME,AGE,ADDRESS,SALARY,CREATED_AT) VALUES (30, 'Saima', 32, 'California', 20000.00, now());
INSERT 0 1

# View data in "company" table
test=# SELECT * FROM company;
 id |  name  | age |                      address                       | salary |         created_at
----+--------+-----+----------------------------------------------------+--------+----------------------------
 10 | Paul   |  32 | California                                         |  20000 | 2026-10-05 09:12:04.123456
 20 | Omkar  |  32 | California                                         |  20000 | 2026-10-05 09:12:10.654321
 30 | Saima  |  32 | California                                         |  20000 | 2026-10-05 09:12:15.987654
(3 rows)
```

## Protect the Application

You can now take a backup of the PostgreSQL data using an ActionSet defining backup for this application. Create an ActionSet in the same namespace as the controller.

```bash
$ kanctl create actionset --action backup --namespace kanister --blueprint pgo-blueprint \
    --objects postgres-operator.crunchydata.com/v1beta1/postgresclusters/pgo-test/hippo
actionset backup-8xk2q created

# View the status of the actionset
$ kubectl --namespace kanister get actionsets.cr.kanister.io backup-8xk2q
NAME           PROGRESS   LAST TRANSITION TIME   STATE
backup-8xk2q   100.00     2026-10-05T09:20:41Z   complete
```

Once the backup completes, the ActionSet's `backupInfo` artifact holds the `pgoBackupRepo`, `pgoBackupTimestamp` and `pgoBackupTimelineID` values that the restore action uses:

```bash
$ kubectl --namespace kanister get actionsets.cr.kanister.io backup-8xk2q \
    -o jsonpath='{.status.actions[0].artifacts.backupInfo.keyValue}'
```

You can also confirm the backup from PGO's side:

```bash
$ kubectl get postgrescluster hippo -n pgo-test -o jsonpath='{.status.pgbackrest.manualBackup}'
```

### Disaster strikes!

Let's say someone accidentally deleted the test database using the following command:

```bash
# Connect to PostgreSQL by running a shell inside the primary database pod
$ kubectl exec -ti -n pgo-test -c database \
    $(kubectl get pods -n pgo-test --selector=postgres-operator.crunchydata.com/cluster=hippo,postgres-operator.crunchydata.com/role=master -o=jsonpath='{.items[0].metadata.name}') -- bash

$ psql

postgres=# \l
                                                     List of databases
   Name    |  Owner   | Encoding | Locale Provider |   Collate   |    Ctype    | ICU Locale | ICU Rules |   Access privileges
-----------+----------+----------+-----------------+-------------+-------------+------------+-----------+-----------------------
 hippo     | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           | =Tc/postgres         +
 postgres  | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           |
 template0 | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           | =c/postgres          +
 template1 | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           | =c/postgres          +
 test      | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           |
(5 rows)

postgres=# DROP DATABASE test;
DROP DATABASE

postgres=# \l
                                                     List of databases
   Name    |  Owner   | Encoding | Locale Provider |   Collate   |    Ctype    | ICU Locale | ICU Rules |   Access privileges
-----------+----------+----------+-----------------+-------------+-------------+------------+-----------+-----------------------
 hippo     | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           | =Tc/postgres         +
 postgres  | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           |
 template0 | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           | =c/postgres          +
 template1 | postgres | UTF8     | libc            | en_US.utf-8 | en_US.utf-8 |            |           | =c/postgres          +
(4 rows)
```

### Restore the Application

To restore the missing data, you should use the backup that you created before. An easy way to do this is to leverage `kanctl`, a command-line tool that helps create ActionSets that depend on other ActionSets:

```bash
# Make sure to use correct backup actionset name here
$ kanctl --namespace kanister create actionset --action restore --from "backup-8xk2q"
actionset restore-backup-8xk2q-6vbbr created

# View the status of the ActionSet
$ kubectl --namespace kanister get actionsets.cr.kanister.io restore-backup-8xk2q-6vbbr
NAME                         PROGRESS   LAST TRANSITION TIME   STATE
restore-backup-8xk2q-6vbbr   100.00     2026-10-05T09:31:17Z   complete
```

**NOTE:**

The restore is an **in-place point-in-time recovery**. PGO stops the cluster's database instances, restores the data to the timestamp recorded during backup, and starts the instances again. Any changes made after the backup are lost.

Once the ActionSet status is set to "complete", you can see that the data has been successfully restored to PostgreSQL

```bash
# Connect to PostgreSQL by running a shell inside the primary database pod
$ kubectl exec -ti -n pgo-test -c database \
    $(kubectl get pods -n pgo-test --selector=postgres-operator.crunchydata.com/cluster=hippo,postgres-operator.crunchydata.com/role=master -o=jsonpath='{.items[0].metadata.name}') -- bash

$ psql

postgres=# \c test
You are now connected to database "test" as user "postgres".

test=# SELECT * FROM company;
 id |  name  | age |                      address                       | salary |         created_at
----+--------+-----+----------------------------------------------------+--------+----------------------------
 10 | Paul   |  32 | California                                         |  20000 | 2026-10-05 09:12:04.123456
 20 | Omkar  |  32 | California                                         |  20000 | 2026-10-05 09:12:10.654321
 30 | Saima  |  32 | California                                         |  20000 | 2026-10-05 09:12:15.987654
(3 rows)
```

### Delete the Artifacts

This blueprint doesn't define a `delete` action. pgBackRest stores the backups in the repository configured on the `PostgresCluster`, and their lifecycle is governed by the pgBackRest retention settings (for example `repo1-retention-full`) in `spec.backups.pgbackrest.global`. See the [PGO retention documentation](https://access.crunchydata.com/documentation/postgres-operator/latest/tutorials/backups-disaster-recovery/backups#managing-backup-retention) for details.

## Troubleshooting

If you run into any issues with the above commands, you can check the logs of the controller using:

```bash
$ kubectl --namespace kanister logs -l app=kanister-operator
```

you can also check events of the actionset

```bash
$ kubectl describe actionset restore-backup-8xk2q-6vbbr -n kanister
```

Backup and restore are carried out by PGO-managed jobs, so their logs and the `PostgresCluster` status are also useful:

```bash
# pgBackRest backup/restore jobs created by PGO
$ kubectl get jobs -n pgo-test

# Backup and restore status reported by PGO
$ kubectl get postgrescluster hippo -n pgo-test -o jsonpath='{.status.pgbackrest}'

# PGO operator logs
$ kubectl logs -n postgres-operator -l postgres-operator.crunchydata.com/control-plane=postgres-operator
```

## Cleanup

### Delete the PostgresCluster

To delete the `hippo` PostgresCluster:

```bash
$ kubectl delete postgrescluster hippo -n pgo-test
```

### Uninstalling PGO

To uninstall the PGO operator:

```bash
$ helm uninstall pgo -n postgres-operator
```

### Delete CRs
Remove Blueprint, ActionSets and the RBAC created for Kanister

```bash
$ kubectl delete blueprints.cr.kanister.io pgo-blueprint -n kanister

$ kubectl --namespace kanister delete actionsets.cr.kanister.io backup-8xk2q restore-backup-8xk2q-6vbbr

$ kubectl delete clusterrolebinding kanister-operator-role-pgo
$ kubectl delete clusterrole kanister-operator-cluster-role-pgo
```
