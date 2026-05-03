package cmd

import (
	"testing"

	"github.com/spf13/viper"
)

func TestBuildConfig(t *testing.T) {
	// Reset viper before tests
	viper.Reset()

	// Mock viper configuration
	viper.Set("filter", "ERROR")
	viper.Set("exclude", "ignore-me")
	viper.Set("regex", "^[0-9]+")
	viper.Set("container", "api-.*")
	viper.Set("since", "5m")
	viper.Set("output", "logs.json")
	viper.Set("tail", "100")
	viper.Set("dedupe", true)
	viper.Set("json", true)

	cfg, err := buildConfig()
	if err != nil {
		t.Fatalf("buildConfig failed: %v", err)
	}

	if cfg.Filter != "ERROR" {
		t.Errorf("expected Filter to be 'ERROR', got '%s'", cfg.Filter)
	}
	if cfg.Exclude != "ignore-me" {
		t.Errorf("expected Exclude to be 'ignore-me', got '%s'", cfg.Exclude)
	}
	if cfg.RegexFilter == nil || cfg.RegexFilter.String() != "^[0-9]+" {
		t.Errorf("expected RegexFilter to be '^[0-9]+', got '%v'", cfg.RegexFilter)
	}
	if cfg.ContainerFilter == nil || cfg.ContainerFilter.String() != "api-.*" {
		t.Errorf("expected ContainerFilter to be 'api-.*', got '%v'", cfg.ContainerFilter)
	}
	if cfg.Since != "5m" {
		t.Errorf("expected Since to be '5m', got '%s'", cfg.Since)
	}
	if cfg.Output != "logs.json" {
		t.Errorf("expected Output to be 'logs.json', got '%s'", cfg.Output)
	}
	if cfg.TailLines != "100" {
		t.Errorf("expected TailLines to be '100', got '%s'", cfg.TailLines)
	}
	if cfg.Deduplicate != true {
		t.Errorf("expected Deduplicate to be true")
	}
	if cfg.JsonOutput != true {
		t.Errorf("expected JsonOutput to be true")
	}
}

func TestBuildConfig_InvalidRegex(t *testing.T) {
	viper.Reset()
	viper.Set("regex", "[invalid")

	_, err := buildConfig()
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestBuildConfig_InvalidContainerRegex(t *testing.T) {
	viper.Reset()
	viper.Set("container", "[invalid")

	_, err := buildConfig()
	if err == nil {
		t.Fatal("expected error for invalid container regex, got nil")
	}
}
