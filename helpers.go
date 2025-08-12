package blueprints

import (
	"fmt"
	"os"
	"path/filepath"

	crv1alpha1 "github.com/kanisterio/kanister/pkg/apis/cr/v1alpha1"
)

// GetBlueprintPath returns the path to a blueprint file given the app name
func GetBlueprintPath(rootPath, appName string) string {
	return filepath.Join(rootPath, appName, fmt.Sprintf("%s-blueprint.yaml", appName))
}

// ListAvailableBlueprints returns a list of available blueprint applications
func ListAvailableBlueprints(rootPath string) ([]string, error) {
	var blueprints []string

	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			appName := entry.Name()
			blueprintPath := GetBlueprintPath(rootPath, appName)

			// Check if the blueprint file exists
			if _, err := os.Stat(blueprintPath); err == nil {
				blueprints = append(blueprints, appName)
			}
		}
	}

	return blueprints, nil
}

// ReadBlueprintByApp reads a blueprint by application name from the blueprints directory
func ReadBlueprintByApp(rootPath, appName string) (*crv1alpha1.Blueprint, error) {
	blueprintPath := GetBlueprintPath(rootPath, appName)
	return ReadFromFile(blueprintPath)
}
