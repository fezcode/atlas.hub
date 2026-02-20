package install

import (
	"os"

	"atlas.hub/internal/model"
	"github.com/fezcode/go-piml"
)

func LoadManifest(path string) ([]model.Tool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m model.Manifest
	if err := piml.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return m.Tools, nil
}
