package blueprints

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
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

// ReadFromFile reads a blueprint from the embedded filesystem
func ReadFromFile(appName string) (*crv1alpha1.Blueprint, error) {
	blueprintPath := fmt.Sprintf("%s/%s-blueprint.yaml", appName, appName)
	log.Printf("Reading blueprint from path: %s", blueprintPath)

	data, err := embeddedBlueprints.ReadFile(blueprintPath)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found for app %s: %w", appName, err)
	}

	return parseBlueprint(data)
}

// ListBlueprints returns a list of all embedded blueprints
func ListBlueprints() ([]string, error) {
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
