package src

import (
	"testing"
)

func TestModuleName(t *testing.T) {
	if ProjectName() != "go-web-common" {
		t.Errorf("Project name `%s` incorrect", ProjectName())
	}
}
