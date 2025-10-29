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

package blueprints

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
)

//go:embed */*-blueprint.yaml
var embeddedBlueprints embed.FS

// ReadFromEmbeddedFile reads a blueprint from the embedded filesystem
func ReadFromEmbeddedFile(blueprintPath string) ([]byte, error) {
	log.Printf("Reading blueprint from path: %s", blueprintPath)

	data, err := embeddedBlueprints.ReadFile(blueprintPath)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found on path %s: %w", blueprintPath, err)
	}

	return data, nil
}

// ReadFromFile parses and returns Blueprint specs
func ReadFromFile(blueprintPath string) ([]byte, error) {
	log.Printf("Reading blueprint from path: %s", blueprintPath)

	data, err := os.ReadFile(blueprintPath)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found on path %s: %w", blueprintPath, err)
	}

	return data, nil
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
