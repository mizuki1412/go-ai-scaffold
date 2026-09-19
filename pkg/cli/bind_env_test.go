package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestLoadConfigExpandsEnvPlaceholders(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(cfg, []byte("llm:\n  apiKey: \"${TF_TEST_KEY}\"\n  other: \"pre-${TF_TEST_KEY}-post\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TF_TEST_KEY", "secret123")

	viper.Reset()
	viper.Set("config", cfg)
	loadConfig()

	if got := viper.GetString("llm.apiKey"); got != "secret123" {
		t.Errorf("llm.apiKey = %q, want secret123", got)
	}
	if got := viper.GetString("llm.other"); got != "pre-secret123-post" {
		t.Errorf("llm.other = %q, want pre-secret123-post", got)
	}
}
