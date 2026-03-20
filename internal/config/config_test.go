package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetEffectiveUserAgentSupportsLegacyOpencode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DisguiseTool = "opencode"

	if got := cfg.GetEffectiveUserAgent(); got != "opencode/0.3.0 (linux)" {
		t.Fatalf("expected legacy opencode user agent, got %q", got)
	}
}

func TestGetEffectiveUserAgentFallsBackToClaudeCode(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DisguiseTool = "unknown-tool"

	if got := cfg.GetEffectiveUserAgent(); got != PredefinedDisguiseTools["claudecode"].UserAgent {
		t.Fatalf("expected Claude Code fallback user agent, got %q", got)
	}
}

func TestFindConfigInDirPrefersConfigToml(t *testing.T) {
	dir := t.TempDir()

	for _, name := range []string{"config.toml", "config.eg", "config.example.toml"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	got, ok := findConfigInDir(dir)
	if !ok {
		t.Fatal("expected config file to be found")
	}

	want := filepath.Join(dir, "config.toml")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestFindConfigInDirFallsBackToConfigEg(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "config.eg"), []byte("example"), 0644); err != nil {
		t.Fatalf("write config.eg: %v", err)
	}

	got, ok := findConfigInDir(dir)
	if !ok {
		t.Fatal("expected config file to be found")
	}

	want := filepath.Join(dir, "config.eg")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
