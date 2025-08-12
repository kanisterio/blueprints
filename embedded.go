package blueprints

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	crv1alpha1 "github.com/kanisterio/kanister/pkg/apis/cr/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

//go:embed */*-blueprint.yaml
var embeddedBlueprints embed.FS

// parseBlueprint parses YAML data into a Blueprint struct
func parseBlueprint(data []byte) (*crv1alpha1.Blueprint, error) {
	var bp crv1alpha1.Blueprint
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1000)
	if err := dec.Decode(&bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

// ReadEmbeddedBlueprint reads a blueprint from the embedded filesystem
func ReadEmbeddedBlueprint(appName string) (*crv1alpha1.Blueprint, error) {
	blueprintPath := fmt.Sprintf("%s/%s-blueprint.yaml", appName, appName)

	data, err := embeddedBlueprints.ReadFile(blueprintPath)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found for app %s: %w", appName, err)
	}

	return parseBlueprint(data)
}

// ListEmbeddedBlueprints returns a list of all embedded blueprints
func ListEmbeddedBlueprints() ([]string, error) {
	var blueprints []string

	err := fs.WalkDir(embeddedBlueprints, ".", func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(filePath, "-blueprint.yaml") {
			// Extract app name from path like "mysql/mysql-blueprint.yaml"
			dir := path.Dir(filePath)
			if dir != "." {
				blueprints = append(blueprints, dir)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return blueprints, nil
}
