// Package utils provides utility functions for blueprint path construction
package utils

import "fmt"

// GetBlueprintPath constructs the path to the blueprint file based on app and blueprint names
func GetBlueprintPathByName(app string, blueprintName string) string {
	var blueprintFolder string

	// If blueprintName is not provided, use app name as blueprint name
	if blueprintName == "" {
		blueprintName = app
	}

	switch app {
	case KanisterAppRDSAuroraSnap:
		blueprintFolder = BPFolderNameAWSRDSAuroraMySQL
	case KanisterAppRDSPostgresSnap, KanisterAppRDSPostgres, KanisterAppRDSPostgresDump:
		blueprintFolder = BPFolderNameAWSRDSPostgres
	case KanisterAppKafka:
		blueprintFolder = BPFolderNameKafkaAdobeS3Connector
	default:
		blueprintFolder = blueprintName
	}

	return fmt.Sprintf("%s/%s-blueprint.yaml", blueprintFolder, blueprintName)
}
