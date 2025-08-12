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

func TestReadBlueprintFromFile(t *testing.T) {
	bp, err := ReadFromFile("mysql")
	if err != nil {
		t.Fatalf("Failed to read MySQL blueprint: %v", err)
	}

	if bp.Kind != "Blueprint" {
		t.Errorf("Expected Kind to be 'Blueprint', got '%s'", bp.Kind)
	}

	if bp.APIVersion == "" {
		t.Error("Expected APIVersion to be set")
	}

	if len(bp.Actions) == 0 {
		t.Error("Expected blueprint to have actions")
	}
}
