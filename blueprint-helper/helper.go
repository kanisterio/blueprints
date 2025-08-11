package blueprinthelper

import (
	"bytes"
	"os"

	crv1alpha1 "github.com/kanisterio/kanister/pkg/apis/cr/v1alpha1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

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
