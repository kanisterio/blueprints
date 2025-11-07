// Copyright 2025 The Kanister Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
