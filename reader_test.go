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
	"slices"
	"testing"
)

func TestListBlueprints(t *testing.T) {
	blueprints, err := ListBlueprints()
	if err != nil {
		t.Fatalf("Failed to list blueprints: %v", err)
	}

	if len(blueprints) == 0 {
		t.Error("Expected at least one blueprint, got none")
	}

	// Check if mysql blueprint exists
	found := slices.Contains(blueprints, "mysql")

	if !found {
		t.Error("Expected to find 'mysql' blueprint in the list")
	}
}

// TODO: add tests for reader functions
