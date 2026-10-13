# Postgres-RDS

[PostgreSQL](https://www.postgresql.org/) is an object-relational database management system (ORDBMS) with an emphasis on extensibility and standards compliance.

## Introduction

This example demonstrates how to use Kanister to protect an AWS RDS PostgreSQL instance using RDS snapshots. The blueprint uses the Kanister functions [CreateRDSSnapshot](https://docs.kanister.io/functions.html#createrdssnapshot), [RestoreRDSSnapshot](https://docs.kanister.io/functions.html#restorerdssnapshot) and [DeleteRDSSnapshot](https://docs.kanister.io/functions.html#deleterdssnapshot) to back up, restore, and delete snapshots of the RDS instance. No custom image is required.

## Prerequisites

- Kubernetes 1.10+
- Kanister controller version 0.118.0 installed in your cluster
- Kanctl CLI installed (https://docs.kanister.io/tooling.html#kanctl)
- AWS credentials with permissions to perform RDS snapshot operations (create, describe, restore and delete DB snapshots and instances)

## Create RDS instance on AWS

> You can skip this step if you already have an RDS instance created

To connect to the RDS instance and populate data, make sure that you have a VPC with a security group that has a rule to allow ingress traffic on 5432 TCP port.

You can create a security group and add rules to the default VPC using the following commands

**NOTE**

This is highly insecure. Its good only for testing and should not be used in a production environment.
For production, make sure to restrict the source (CIDR or otherwise) to only the elements that need to access the DB.

```bash
aws ec2 create-security-group --group-name <security-group-name> --description "pgtest security group"
aws ec2 authorize-security-group-ingress --group-name <security-group-name> --protocol tcp --port 5432 --cidr 0.0.0.0/0
```

Fetch the Security Group ID
``` bash
$ aws ec2 describe-security-groups --filters "Name=group-name,Values=<security-group-name>" --query "SecurityGroups[*].GroupId"
```

Now create an RDS instance with the PostgreSQL engine

```bash
aws rds create-db-instance \
    --publicly-accessible \
    --allocated-storage 20 --db-instance-class db.t3.micro \
    --db-instance-identifier <instance-name> \
    --engine postgres \
    --master-username <master-username> \
    --vpc-security-group-ids <security-group-id> \
    --master-user-password <db-password>

aws rds wait db-instance-available --db-instance-identifier=<instance-name>
```

Where `<security-group-id>` is the security group with the TCP 5432 inbound rule.

## Create configmap

The blueprint is run against a ConfigMap that holds the RDS instance identifier. Create the following ConfigMap e.g. `dbconfig.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: dbconfig
  namespace: postgres-test
data:
  postgres.instanceid: <instance-name>  # identifier of the RDS instance to protect
```

```bash
$ kubectl create namespace postgres-test
$ kubectl apply -f dbconfig.yaml
```

## Integrating with Kanister

If you have created the ConfigMap with a name other than `dbconfig` or in a namespace other than `postgres-test`, you need to modify the commands used below to use the correct name and namespace.

### Create Profile

Create Profile CR if not created already

```bash
$ kanctl create profile s3compliant --access-key <aws-access-key-id> \
	--secret-key <aws-secret-key> \
	--bucket <s3-bucket-name> --region <region-name> \
	--namespace postgres-test
```

**NOTE:**

The blueprint uses the credentials and the region from this Profile to perform the RDS operations, so `<region-name>` must be the region of the RDS instance and the credentials must have proper permissions to perform operations on RDS.
Every ActionSet requires a Profile reference to complete the action. This CR (`profiles.cr.kanister.io`)
can be shared between Kanister-enabled application instances.

### Create Blueprint

Create a Blueprint in the same namespace as the controller.

```bash
$ kubectl create -f ./postgres-rds-restore-blueprint.yaml -n kanister
blueprint.cr.kanister.io/rds-postgres-bp created
```

Once the RDS instance is available, you can populate it with some data. Let's add a table called "company" to a "test" database:

```bash
$ kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace postgres-test --image docker.io/bitnamilegacy/postgresql:17.6.0 --env="PGPASSWORD=<db-password>" \
      --command -- psql --host <rds-instance-endpoint> -U <master-username> -d postgres -p 5432

postgres=> CREATE DATABASE test;
CREATE DATABASE

## Create table COMPANY in test database
postgres=> \c test
test=> CREATE TABLE COMPANY(
     ID INT PRIMARY KEY     NOT NULL,
     NAME           TEXT    NOT NULL,
     AGE            INT     NOT NULL,
     ADDRESS        CHAR(50),
     SALARY         REAL,
     CREATED_AT    TIMESTAMP
);
CREATE TABLE

## Insert data into the table
test=> INSERT INTO COMPANY (ID,NAME,AGE,ADDRESS,SALARY,CREATED_AT) VALUES (10, 'Paul', 32, 'California', 20000.00, now());
INSERT 0 1
test=> INSERT INTO COMPANY (ID,NAME,AGE,ADDRESS,SALARY,CREATED_AT) VALUES (20, 'Omkar', 32, 'California', 20000.00, now());
INSERT 0 1
test=> INSERT INTO COMPANY (ID,NAME,AGE,ADDRESS,SALARY,CREATED_AT) VALUES (30, 'Prasad', 32, 'California', 20000.00, now());
INSERT 0 1

test=> select * from company;
 id |  name  | age |                      address                       | salary |         created_at
----+--------+-----+----------------------------------------------------+--------+----------------------------
 10 | Paul   |  32 | California                                         |  20000 | 2019-09-16 14:39:36.316065
 20 | Omkar  |  32 | California                                         |  20000 | 2019-09-16 14:40:52.952459
 30 | Prasad |  32 | California                                         |  20000 | 2019-09-16 14:41:06.433487
(3 rows)
```

Where `<rds-instance-endpoint>` is the endpoint of the RDS instance, which can be fetched using:

```bash
$ aws rds describe-db-instances --db-instance-identifier <instance-name> --query "DBInstances[0].Endpoint.Address" --output text
```

## Protect the Application

You can now take a snapshot of the RDS instance by creating a backup ActionSet for this application. Create an ActionSet in the same namespace as the controller.

```bash
$ kubectl get profile -n postgres-test
NAME               AGE
s3-profile-7d6wt   7m25s
```

Create an ActionSet file e.g. `rds-backup-actionset.yaml`:

```yaml
apiVersion: cr.kanister.io/v1alpha1
kind: ActionSet
metadata:
  name: rds-backup
  namespace: kanister
spec:
  actions:
  - name: backup
    blueprint: rds-postgres-bp
    object:
      apiVersion: v1
      kind: ConfigMap
      name: dbconfig
      namespace: postgres-test
      resource: configmaps
    profile:
      name: s3-profile-7d6wt
      namespace: postgres-test
```

```bash
$ kubectl apply -f rds-backup-actionset.yaml
actionset.cr.kanister.io/rds-backup created

# View the status of the actionset
$ kubectl --namespace kanister describe actionset rds-backup
```

Once the ActionSet is complete, the RDS snapshot can be seen using:

```bash
$ aws rds describe-db-snapshots --db-instance-identifier <instance-name> --region <region-name>
```

### Disaster strikes!

Let's say someone accidentally deleted the `company` table:

```bash
$ kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace postgres-test --image docker.io/bitnamilegacy/postgresql:17.6.0 --env="PGPASSWORD=<db-password>" \
      --command -- psql --host <rds-instance-endpoint> -U <master-username> -d test -p 5432

test=> DROP TABLE company;
DROP TABLE
```

### Restore the Application

To restore the RDS instance from the snapshot, you should use the backup that you created before. An easy way to do this is to leverage `kanctl`, a command-line tool that helps create ActionSets that depend on other ActionSets:

```bash
$ kanctl --namespace kanister create actionset --action restore --from rds-backup
actionset restore-rds-backup-6jzt4 created

## Check status
$ kubectl --namespace kanister describe actionset restore-rds-backup-6jzt4
```

**NOTE:**

The restore action replaces the existing RDS instance: the instance with the identifier stored in the backup artifact is deleted and a new instance with the same identifier is restored from the snapshot, using the security group and DB subnet group captured at backup time. The instance is unavailable during the restore.

Once the ActionSet status is set to "complete", you can verify that the data has been restored by connecting to the RDS instance:

```bash
$ kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace postgres-test --image docker.io/bitnamilegacy/postgresql:17.6.0 --env="PGPASSWORD=<db-password>" \
      --command -- psql --host <rds-instance-endpoint> -U <master-username> -d test -p 5432

test=> select * from company;
 id |  name  | age |                      address                       | salary |         created_at
----+--------+-----+----------------------------------------------------+--------+----------------------------
 10 | Paul   |  32 | California                                         |  20000 | 2019-09-16 14:39:36.316065
 20 | Omkar  |  32 | California                                         |  20000 | 2019-09-16 14:40:52.952459
 30 | Prasad |  32 | California                                         |  20000 | 2019-09-16 14:41:06.433487
(3 rows)
```

### Delete the Artifacts

The RDS snapshot created by the backup action can be deleted using the following command:

```bash
$ kanctl --namespace kanister create actionset --action delete --from rds-backup --namespacetargets kanister
actionset delete-rds-backup-cq6bw created

# View the status of the ActionSet
$ kubectl --namespace kanister describe actionset delete-rds-backup-cq6bw
```

## Troubleshooting

If you run into any issues with the above commands, you can check the logs of the controller using:

```bash
$ kubectl --namespace kanister logs -l app=kanister-operator
```

you can also check the events of the actionset

```bash
$ kubectl describe actionset <actionset-name> -n kanister
```

## Cleanup

### Delete the RDS instance

```bash
$ aws rds delete-db-instance --db-instance-identifier <instance-name> --skip-final-snapshot
```

### Delete CRs
Remove Blueprint, Profile CR and ConfigMap

```bash
$ kubectl delete blueprints.cr.kanister.io rds-postgres-bp -n kanister

$ kubectl get profiles.cr.kanister.io -n postgres-test
NAME               AGE
s3-profile-7d6wt   17m
$ kubectl delete profiles.cr.kanister.io s3-profile-7d6wt -n postgres-test

$ kubectl delete configmap dbconfig -n postgres-test
```
