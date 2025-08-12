package main

import (
	"fmt"
	"log"

	"github.com/kanisterio/blueprints"
)

func main() {
	fmt.Println("=== Kanister Blueprints Go Module Demo ===")

	// Example 1: List all embedded blueprints
	fmt.Println("\n1. Listing all available blueprints:")
	apps, err := blueprints.ListEmbeddedBlueprints()
	if err != nil {
		log.Printf("Error listing blueprints: %v", err)
	} else {
		for i, app := range apps {
			fmt.Printf("   %d. %s\n", i+1, app)
		}
	}

	// Example 2: Read a specific blueprint
	fmt.Println("\n2. Reading MySQL blueprint:")
	bp, err := blueprints.ReadEmbeddedBlueprint("mysql")
	if err != nil {
		log.Printf("Error reading MySQL blueprint: %v", err)
	} else {
		fmt.Printf("   API Version: %s\n", bp.APIVersion)
		fmt.Printf("   Kind: %s\n", bp.Kind)
		fmt.Printf("   Name: %s\n", bp.ObjectMeta.Name)
		fmt.Printf("   Number of Actions: %d\n", len(bp.Actions))

		// List action names
		actionNames := make([]string, 0, len(bp.Actions))
		for name := range bp.Actions {
			actionNames = append(actionNames, name)
		}
		fmt.Printf("   Action Names: %v\n", actionNames)
	}

	// Example 3: Demonstrate file system access
	fmt.Println("\n3. File system access example:")
	fmt.Println("   To use file system access in other repos:")
	fmt.Println("   rootPath := \"/path/to/blueprints\"")
	fmt.Println("   bp, err := blueprints.ReadBlueprintByApp(rootPath, \"mysql\")")
	fmt.Println("   // or")
	fmt.Println("   bp, err := blueprints.ReadFromFile(\"/path/to/mysql/mysql-blueprint.yaml\")")
}
