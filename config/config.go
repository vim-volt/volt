package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/vim-volt/volt/pathutil"
)

type Config struct {
	Build ConfigBuild `toml:"build"`
}

type ConfigBuild struct {
	Strategy string `toml:"strategy"`
}

const (
	SymlinkBuilder = "symlink"
	CopyBuilder    = "copy"
)

func initialConfigTOML() *Config {
	return &Config{
		ConfigBuild{
			Strategy: SymlinkBuilder,
		},
	}
}

func Read() (*Config, error) {
	// Return initial lock.json struct if lockfile does not exist
	configFile := pathutil.ConfigTOML()
	initCfg := initialConfigTOML()
	if !pathutil.Exists(configFile) {
		return initCfg, nil
	}

	var cfg Config
	if _, err := toml.DecodeFile(configFile, &cfg); err != nil {
		return nil, err
	}
	merge(&cfg, initCfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func merge(cfg, initCfg *Config) {
	if cfg.Build.Strategy == "" {
		cfg.Build.Strategy = initCfg.Build.Strategy
	}
}

func validate(cfg *Config) error {
	if cfg.Build.Strategy != "symlink" && cfg.Build.Strategy != "copy" {
		return fmt.Errorf("build.strategy is %q: valid values are %q or %q", cfg.Build.Strategy, "symlink", "copy")
	}
	return nil
}
