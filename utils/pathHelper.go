// Package utils provides utility functions for blueprint path construction
package utils

import "fmt"

// GetBlueprintPathByName constructs the path to the blueprint file based on app and blueprint names
func GetBlueprintPathByName(app string, blueprintName string) string {
	var blueprintFolder string

	// If blueprintName is not provided, use app name as blueprint name
	if blueprintName == "" {
		blueprintName = app
	}

	switch app {
	case kanisterAppRDSAuroraSnap:
		blueprintFolder = bpFolderNameAWSRDSAuroraMySQL
	case kanisterAppRDSPostgresSnap, kanisterAppRDSPostgres, kanisterAppRDSPostgresDump:
		blueprintFolder = bpFolderNameAWSRDSPostgres
	case kanisterAppKafka:
		blueprintFolder = bpFolderNameKafkaAdobeS3Connector
	default:
		blueprintFolder = blueprintName
	}

	return fmt.Sprintf(blueprintPathFormat, blueprintFolder, blueprintName)
}
