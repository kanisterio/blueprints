# Kanister Blueprints

This repository provides a curated set of [Kanister](https://kanister.io) blueprints to help protect a variety of applications and databases on Kubernetes. These examples are designed to make experimenting with Kanister easier.

## 🚀 What are Kanister Blueprints?

Blueprints are custom resources that define how to perform data operations (backup, restore, delete) for specific applications. They contain the business logic needed to safely and consistently protect your application data.

## 📋 Prerequisites

- Kubernetes 1.20+
- [Kanister controller](https://docs.kanister.io/install.html) version 0.114.0+ installed in your cluster
- [Kanctl CLI](https://docs.kanister.io/tooling.html#install-the-tools) installed
- Helm for deploying applications

## 🎯 Getting Started

If you're new to Kanister, we recommend starting with the **[time-log](./time-log/)** blueprint. This simple example demonstrates the core concepts of Kanister without the complexity of a real database.

## 📚 Available Blueprints

- **[AWS RDS PostgreSQL](./aws-rds-postgres/aws-rds-postgres-blueprint.yaml)** 
- **[Cassandra](./cassandra/cassandra-blueprint.yaml)** 
- **[CockroachDB](./cockroachdb/cockroachdb-blueprint.yaml)** 
- **[CSI Snapshots](./csi-snapshot/csi-snapshot-blueprint.yaml)** 
- **[Couchbase](./couchbase/couchbase-blueprint.yaml)** 
- **[Elasticsearch](./elasticsearch/elasticsearch-blueprint.yaml)** 
- **[etcd (OpenShift)](./etcd-incluster-ocp/etcd-incluster-ocp-blueprint.yaml)** 
- **[FoundationDB](./foundationdb/foundationdb-blueprint.yaml)** 
- **[K8ssandra](./k8ssandra/k8ssandra-blueprint.yaml)** 
- **[MSSQL Server](./mssql/mssql-blueprint.yaml)**
- **[MySQL](./mysql/mysql-blueprint.yaml)** 
- **[MySQL on OpenShift using DeploymentConfig](./mysql-dep-config/)** 
- **[MongoDB](./mongodb/mongodb-blueprint.yaml)** 
- **[MongoDB on OpenShift using DeploymentConfig](./mongodb-dep-config/mongodb-dep-config-blueprint.yaml)** 
- **[MongoDB Atlas](./mongodb-atlas/mongodb-atlas-blueprint.yaml)**  
- **[PostgreSQL](./postgres/postgres-blueprint.yaml)** 
- **[PostgreSQL on OpenShift using DeploymentConfig](./postgres-dep-config/postgres-dep-config-blueprint.yaml)**
- **[Redis](./redis/redis-blueprint.yaml)** 


## Getting Help

If you have any questions or run into issues, feel free to reach out to us on
[Slack](https://kanisterio.slack.com).

## License

Apache License 2.0. See [LICENSE](https://github.com/kanisterio/kanister/blob/master/LICENSE).

---

**Note**: Each blueprint includes detailed setup instructions in its respective `README.md` file. Make sure to follow the specific requirements and configurations for your chosen application.
