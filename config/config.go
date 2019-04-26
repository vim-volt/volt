package config

import (
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/vim-volt/volt/pathutil"
)

// Config is marshallable content of config.toml
type Config struct {
	Alias map[string][]string `toml:"alias"`
	Build configBuild         `toml:"build"`
	Get   configGet           `toml:"get"`
	Edit  configEdit          `toml:"edit"`
}

// configBuild is a config for 'volt build'.
type configBuild struct {
	Strategy string `toml:"strategy"`
}

// configGet is a config for 'volt get'.
type configGet struct {
	CreateSkeletonPlugconf *bool `toml:"create_skeleton_plugconf"`
	FallbackGitCmd         *bool `toml:"fallback_git_cmd"`
}

// configEdit is a config for 'volt edit'.
type configEdit struct {
	Editor string `toml:"editor"`
}

const (
	// SymlinkBuilder creates symlinks when 'volt build'.
	SymlinkBuilder = "symlink"
	// CopyBuilder copies/creates regular files when 'volt build'.
	CopyBuilder = "copy"
)

func initialConfigTOML() *Config {
	trueValue := true
	falseValue := false
	return &Config{
		Build: configBuild{
			Strategy: SymlinkBuilder,
		},
		Get: configGet{
			CreateSkeletonPlugconf: &trueValue,
			FallbackGitCmd:         &falseValue,
		},
		Edit: configEdit{
			Editor: "",
		},
	}
}

// Read reads from config.toml and returns Config
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
	if cfg.Get.CreateSkeletonPlugconf == nil {
		cfg.Get.CreateSkeletonPlugconf = initCfg.Get.CreateSkeletonPlugconf
	}
	if cfg.Get.FallbackGitCmd == nil {
		cfg.Get.FallbackGitCmd = initCfg.Get.FallbackGitCmd
	}
	if cfg.Edit.Editor == "" {
		cfg.Edit.Editor = initCfg.Edit.Editor
	}
}

func validate(cfg *Config) error {
	if cfg.Build.Strategy != "symlink" && cfg.Build.Strategy != "copy" {
		return errors.Errorf("build.strategy is %q: valid values are %q or %q", cfg.Build.Strategy, "symlink", "copy")
	}
	return nil
}
