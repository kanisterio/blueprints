package blueprinthelper

import (
	"bytes"
	"embed"
	"os"

	crv1alpha1 "github.com/kanisterio/kanister/pkg/apis/cr/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

var blueprintFiles embed.FS

// ReadFromFile parsed and returns Blueprint specs placed at blueprints/{app}-blueprint.yaml
func ReadFromFile(path string) (*crv1alpha1.Blueprint, error) {
	bpRaw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bp crv1alpha1.Blueprint
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(bpRaw), 1000)
	if err := dec.Decode(&bp); err != nil {
		return nil, err
	}

	return &bp, err
}

// ReadFromEmbedded reads and returns Blueprint specs from embedded files
// path should be relative to the blueprints directory (e.g., "mysql/mysql-blueprint.yaml")
func ReadFromEmbedded(path string) (*crv1alpha1.Blueprint, error) {
	bpRaw, err := blueprintFiles.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var bp crv1alpha1.Blueprint
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(bpRaw), 1000)
	if err := dec.Decode(&bp); err != nil {
		return nil, err
	}

	return &bp, nil
}

// ListEmbeddedBlueprints returns a list of all embedded blueprint file paths
func ListEmbeddedBlueprints() ([]string, error) {
	var blueprints []string

	err := walkEmbeddedDir(".", &blueprints)
	if err != nil {
		return nil, err
	}

	return blueprints, nil
}

// walkEmbeddedDir recursively walks the embedded filesystem to find blueprint files
func walkEmbeddedDir(dir string, blueprints *[]string) error {
	entries, err := blueprintFiles.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := dir + "/" + entry.Name()
		if dir == "." {
			path = entry.Name()
		}

		if entry.IsDir() {
			err := walkEmbeddedDir(path, blueprints)
			if err != nil {
				continue // Skip directories that can't be read
			}
		} else if entry.Type().IsRegular() {
			// Check if it's a blueprint file
			name := entry.Name()
			if (len(name) > 9 && name[len(name)-9:] == "blueprint.yaml") ||
				(len(name) > 14 && name[len(name)-14:] == "-blueprint.yaml") {
				*blueprints = append(*blueprints, path)
			}
		}
	}

	return nil
}
