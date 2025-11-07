package utils

// kanister App Names
const (
	kanisterAppRDSAuroraSnap   = "rds-aurora-snap"
	kanisterAppRDSPostgresSnap = "rds-postgres-snap"
	kanisterAppRDSPostgres     = "rds-postgres"
	kanisterAppRDSPostgresDump = "rds-postgres-dump"
	kanisterAppKafka           = "kafka"
)

// Blueprint Folder Names
const (
	bpFolderNameAWSRDSAuroraMySQL     = "aws-rds-aurora-mysql"
	bpFolderNameAWSRDSPostgres        = "aws-rds-postgres"
	bpFolderNameKafkaAdobeS3Connector = "kafka-adobe-s3-connector"
)

const blueprintPathFormat = "%s/%s-blueprint.yaml"
