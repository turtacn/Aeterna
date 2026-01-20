package cli

import (
	"testing"
)

func TestCommands(t *testing.T) {
	if rootCmd.Name() != "aeterna" {
		t.Errorf("Expected root command name aeterna, got %s", rootCmd.Name())
	}

	if len(rootCmd.Commands()) < 2 {
		t.Errorf("Expected at least 2 subcommands, got %d", len(rootCmd.Commands()))
	}
}
