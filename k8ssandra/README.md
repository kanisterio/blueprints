# K8ssandra

[K8ssandra](https://k8ssandra.io/) is a cloud-native distribution of Apache Cassandra (Cassandra) designed to run on Kubernetes. K8ssandra follows the K8s operator pattern to automate operational tasks. This includes metric, data anti-entropy services, and backup/restore tooling. More details can be found [here](https://docs.k8ssandra.io).

K8ssandra operator uses Medusa to backup and restore Cassandra data. Kanister can make use of Medusa operator APIs to perform backup and restore of Cassandra data.

## Prerequisites

* Kubernetes 1.17+
* PV support on the underlying infrastructure
* Kanister controller version 0.119.0 installed in your cluster
* Kanctl CLI installed (https://docs.kanister.io/tooling.html#kanctl)
* K8ssandra needs at least 4 cores and 8GB of RAM available to Docker and appropriate heap sizes for Cassandra and Stargate. If you don’t have those resources available, you can avoid deploying features such as monitoring, Reaper and Medusa, and also reduce the number of Cassandra nodes.

## Chart Details

We will be using the official [k8ssandra-operator](https://github.com/k8ssandra/k8ssandra-operator) Helm chart to deploy the K8ssandra stack on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

The blueprint uses the Medusa APIs (`medusa.k8ssandra.io/v1alpha1` - `MedusaBackupJob`, `MedusaBackup` and `MedusaRestoreJob`) that are provided by k8ssandra-operator. The legacy `k8ssandra/k8ssandra` (v1) chart and its `cassandra.k8ssandra.io` `CassandraBackup`/`CassandraRestore` APIs are not supported by this blueprint.

## Installing the K8ssandra

> Complete guide to install k8ssandra-operator can be found [here](https://docs.k8ssandra.io/install/).

Before we begin, make sure that there exists a storageclass with **VOLUMEBINDINGMODE of WaitForFirstConsumer.**

### Install cert-manager and k8ssandra-operator

k8ssandra-operator requires [cert-manager](https://cert-manager.io) to be installed in the cluster.

```bash
# Install cert-manager
$ helm repo add jetstack https://charts.jetstack.io
$ helm repo update
$ helm install cert-manager jetstack/cert-manager -n cert-manager --create-namespace --set crds.enabled=true

# Add k8ssandra Helm repo
$ helm repo add k8ssandra https://helm.k8ssandra.io/stable
$ helm repo update

# Install k8ssandra-operator
$ helm install k8ssandra-operator k8ssandra/k8ssandra-operator -n k8ssandra --create-namespace
```

### Create S3 creds secret for Medusa Amazon S3 integration

K8ssandra Operator deploys Medusa to support backup and restore operations of Apache Cassandra® tables.

Medusa supports different types of object stores including GCS, S3, and S3-compatible stores like MinIO, CEPH Object Gateway, etc. The complete list can be found [here](https://docs.k8ssandra.io/tasks/backup-restore/).

For the scope of this example, we will just focus on S3 bucket. But you can choose any object store of your choice.

Create a secret in the `k8ssandra` namespace using the following template to enable S3 integration with Medusa. Replace `my_access_key` and `my_secret_key` with actual keys before creating the secret.

```yaml
apiVersion: v1
kind: Secret
metadata:
 name: medusa-bucket-key
 namespace: k8ssandra
type: Opaque
stringData:
 # Note that this currently has to be set to credentials!
 credentials: |-
   [default]
   aws_access_key_id = my_access_key
   aws_secret_access_key = my_secret_key
```

### Create K8ssandra cluster

Create a `K8ssandraCluster` resource e.g. k8ssandra-cluster.yaml with the following configuration so that K8ssandra installation works on the limited resources we have locally.

```yaml
apiVersion: k8ssandra.io/v1alpha1
kind: K8ssandraCluster
metadata:
  name: k8ssandra
  namespace: k8ssandra
spec:
  cassandra:
    serverVersion: "4.0.1"
    datacenters:
      - metadata:
          name: dc1
        size: 1
        storageConfig:
          cassandraDataVolumeClaimSpec:
            storageClassName: <storage_class>
            accessModes:
              - ReadWriteOnce
            resources:
              requests:
                storage: 5Gi
        config:
          jvmOptions:
            heapSize: 512M
  medusa:
    storageProperties:
      storageProvider: s3
      storageSecretRef:
        name: medusa-bucket-key
      bucketName: <aws_s3_bucket_name>
      region: <aws_bucket_region>
      prefix: k8ssandra
      secure: true
```

Where,

- `storage_class` is the storageclass with `WaitForFirstConsumer` VolumeBindingMode.
- `aws_s3_bucket_name` is the S3 bucket that will be used for backup by Medusa.
- `aws_bucket_region` is the AWS region bucket exists.

```bash
$ kubectl apply -f k8ssandra-cluster.yaml
```

The operator will create a `CassandraDatacenter` named `dc1` in the `k8ssandra` namespace. The Cassandra pod is named `k8ssandra-dc1-default-sts-0`, which we’ll use throughout the following example. Wait till the datacenter is ready:

```bash
$ kubectl wait cassandradatacenters.cassandra.datastax.com/dc1 -n k8ssandra --for=condition=Ready --timeout=15m
```


**NOTE:**

Kanister operator role needs to be updated in order to access the `CassandraDatacenter` and Medusa (`medusa.k8ssandra.io`) custom resources.
Create required ClusterRole and ClusterRoleBinding with the following command.

```bash
$ kubectl auth reconcile -f kanister-k8ssandra-rbac.yaml
```

You may have to update the service account name in the specs if you have deployed Kanister with another name

## Integrating with Kanister

If you have deployed the K8ssandra operator with other name than `k8ssandra` and namespace other than `k8ssandra`, you need to modify the commands used below to use the correct name and namespace.

### Create Blueprint

Create Blueprint in the same namespace as the controller

```bash
$ kubectl create -f k8ssandra-blueprint.yaml -n kanister
blueprint.cr.kanister.io/k8ssandra-blueprint created
```

Once K8ssandra pods are running, you can populate it with some data.

```bash
# k8ssandra-operator enables authentication by default, fetch the superuser credentials
$ CASS_USERNAME=$(kubectl get secret k8ssandra-superuser -n k8ssandra -o jsonpath='{.data.username}' | base64 --decode)
$ CASS_PASSWORD=$(kubectl get secret k8ssandra-superuser -n k8ssandra -o jsonpath='{.data.password}' | base64 --decode)

# Connect to the Cassandra CLI (cqlsh) in the cassandra pod
$ kubectl exec -ti k8ssandra-dc1-default-sts-0 -c cassandra -n k8ssandra -- cqlsh -u $CASS_USERNAME -p $CASS_PASSWORD

# create the keyspace
$ cqlsh> create keyspace restaurants with replication  = {'class':'SimpleStrategy', 'replication_factor': 1};

# once the keyspace is created let's create a table named guests and some data into that table
$ cqlsh> create table restaurants.guests (id UUID primary key, firstname text, lastname text, birthday timestamp);
$ cqlsh> insert into restaurants.guests (id, firstname, lastname, birthday)  values (5b6962dd-3f90-4c93-8f61-eabfa4a803e2, 'Vivek', 'Singh', '2015-02-18');
$ cqlsh> insert into restaurants.guests (id, firstname, lastname, birthday)  values (5b6962dd-3f90-4c93-8f61-eabfa4a803e4, 'Prasad', 'Hemsworth', '2015-02-18');
$ cqlsh> insert into restaurants.guests (id, firstname, lastname, birthday)  values (5b6962dd-3f90-4c93-8f61-eabfa4a803e3, 'Tom', 'Singh', '2015-02-18');

# verify data
$ cqlsh> select * from restaurants.guests;

 id                                   | birthday                        | firstname | lastname
--------------------------------------+---------------------------------+-----------+-----------
 5b6962dd-3f90-4c93-8f61-eabfa4a803e2 | 2015-02-18 00:00:00.000000+0000 |     Vivek |     Singh
 5b6962dd-3f90-4c93-8f61-eabfa4a803e3 | 2015-02-18 00:00:00.000000+0000 |       Tom |     Singh
 5b6962dd-3f90-4c93-8f61-eabfa4a803e4 | 2015-02-18 00:00:00.000000+0000 |    Prasad | Hemsworth

(3 rows)

```

## Protect the Application

You can now take a backup of the K8ssandra cluster data using an ActionSet defining backup for this application. Create an ActionSet in the same namespace as the controller.

```bash
# Create actionset to run backup action from the blueprint
$ kanctl create actionset --action backup --blueprint k8ssandra-blueprint --objects cassandra.datastax.com/v1beta1/cassandradatacenters/k8ssandra/dc1 -n kanister
actionset backup-swxnq created

# you can check the status of the actionset either by describing the actionset resource or by checking the kanister operator's pod log
$ kubectl describe actionset backup-swxnq -n kanister

# Wait till the actionset `completes`

# The backup action creates a MedusaBackupJob and waits for the corresponding MedusaBackup to finish
$ kubectl get medusabackupjobs.medusa.k8ssandra.io,medusabackups.medusa.k8ssandra.io -n k8ssandra
```

### Disaster strikes!

Let's say someone accidentally deleted the keyspace using the following command in K8ssandra cluster pod.

```bash
# Connect to the Cassandra CLI (cqlsh) in the cassandra pod
$ kubectl exec -ti k8ssandra-dc1-default-sts-0 -c cassandra -n k8ssandra -- cqlsh -u $CASS_USERNAME -p $CASS_PASSWORD

# drop the guests table
$ cqlsh> drop table if exists restaurants.guests;
# drop restaurants keyspace
$ cqlsh> drop  keyspace  restaurants;
```

### Restore the Application

Now that we have removed the data from the Cassandra database's table, let's restore the data using the backup that we have already created in the earlier step. To do that we will again create an Actionset resource but for restore instead of backup. You can create the Actionset using the below command

```bash
$ kanctl create actionset --action restore --from backup-swxnq -n kanister
actionset restore-backup-swxnq-9hhmd created

# View the status of the ActionSet
$ kubectl describe actionset restore-backup-swxnq-9hhmd --namespace kanister
```

The restore action creates a `MedusaRestoreJob`, waits for it to finish and then deletes it. Medusa restores the datacenter in place, so the Cassandra pods are restarted during the restore.

Once you have verified that the status of the Actionset <restore-actionset> is completed. You can check if the data is restored or not by EXECing into the Cassandra pod and selecting all the data from the table.

```bash
$ kubectl exec -ti k8ssandra-dc1-default-sts-0 -c cassandra -n k8ssandra -- cqlsh -u $CASS_USERNAME -p $CASS_PASSWORD
$ cqlsh> select * from restaurants.guests;
```

### Delete the Artifacts

The artifacts created by the backup action can be cleaned up using the following command. The delete action removes both the `MedusaBackup` and the `MedusaBackupJob` resources created by the backup action.

```bash
$ kanctl create actionset --action delete --from backup-swxnq -n kanister --namespacetargets kanister
actionset delete-backup-swxnq-l5zc4 created


# View the status of the ActionSet
$ kubectl --namespace kanister describe actionset delete-backup-swxnq-l5zc4
```

### Troubleshooting

If you run into any issues with the above commands, you can check the logs of the controller using.

```bash
$ kubectl --namespace <kanister-operator-namespace> logs -l app=kanister-operator
```

You can also check events of the ActionSet with the following command.

```bash
$ kubectl describe actionset <actionset-name> -n <kanister-operator-namespace>
```

## Uninstalling the Chart

To uninstall/delete the K8ssandra application run the below commands.

```bash
$ kubectl delete k8ssandraclusters.k8ssandra.io k8ssandra -n k8ssandra
$ helm delete k8ssandra-operator -n k8ssandra
```

The commands remove the K8ssandra cluster, all the Kubernetes components associated with the chart and delete the release.

## Delete Blueprint and Profile CR

```bash
$ kubectl delete blueprints.cr.kanister.io k8ssandra-blueprint -n kanister
```
