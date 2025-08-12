package blueprints

import (
	"path/filepath"
	"testing"
)

func TestReadFromFile(t *testing.T) {
	// Test reading a real blueprint file
	blueprintPath := filepath.Join("mysql", "mysql-blueprint.yaml")

	bp, err := ReadFromFile(blueprintPath)
	if err != nil {
		t.Fatalf("Failed to read blueprint: %v", err)
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

func TestGetBlueprintPath(t *testing.T) {
	path := GetBlueprintPath("/root", "mysql")
	expected := filepath.Join("/root", "mysql", "mysql-blueprint.yaml")

	if path != expected {
		t.Errorf("Expected path '%s', got '%s'", expected, path)
	}
}
