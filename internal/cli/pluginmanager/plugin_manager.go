package pluginmanager

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Reader interface {
	// data contains the container for which data will be loaded
	ReadFile(location string, data any) error
	ReadIntoRawBytes(location string) ([]byte, error)
}

// Builtin will packaged as zip
// Installed on runtime
type PluginManager struct {
	Rules     map[string]*PluginRule
	Reader    Reader
	ApiType   string
	IsDevMode bool
}

// these are the
type PluginRule struct {
	File    string
	Disable bool
	Options map[string]any
}

type PluginUserOverride struct {
	Disable *bool          `json:"omitempty" yaml:"omitempty" toml:"omitempty"`
	Options map[string]any `json:"omitempty" yaml:"omitempty" toml:"omitempty"`
}

type PluginConfFile struct {
	Rules map[string]PluginRule
}

func New(fr Reader, apiType string, isDevMode bool) *PluginManager {
	return &PluginManager{
		Rules:     make(map[string]*PluginRule),
		Reader:    fr,
		ApiType:   apiType,
		IsDevMode: isDevMode,
	}
}

func getPluginConfFile(files []fs.DirEntry) (string, error) {
	for _, file := range files {
		if !file.IsDir() && file.Name()[0:6] == "config" {
			return file.Name(), nil
		}
	}
	return "", errors.New("config file not found")
}

func (p *PluginManager) LoadBuiltinPlugin() error {
	cwd, _ := os.Getwd()
	var path string

	if p.IsDevMode {
		path = filepath.Clean(filepath.Join(cwd, fmt.Sprintf("./plugins/builtin/%s", p.ApiType)))
	} else {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Clean(filepath.Join(homeDir, fmt.Sprintf(".apic/plugins/builtin/%s", p.ApiType)))
	}

	builtInPlugin, err := os.ReadDir(path)
	if err != nil {
		log.Fatal("Failed to open builtin plugins dir: ", err)
	}

	// get plugin config file.
	pluginCfgName, err := getPluginConfFile(builtInPlugin)
	if err != nil {
		return err
	}

	// load plugin config
	cfgFilePath := filepath.Join(path, pluginCfgName)
	var pluginCfg PluginConfFile
	if err := p.Reader.ReadFile(cfgFilePath, &pluginCfg); err != nil {
		return err
	}

	// load up the rules
	for rule, conf := range pluginCfg.Rules {
		jsRuleFile := filepath.Join(path, fmt.Sprintf("/%s", conf.File))
		p.Rules[rule] = &PluginRule{Disable: conf.Disable, File: jsRuleFile, Options: conf.Options}
	}

	return nil
}

func (p *PluginManager) LoadUserPlugins(userPlugins PluginConfFile) error {
	// load up the rules
	for rule, conf := range userPlugins.Rules {
		if _, ok := p.Rules[rule]; ok {
			fmt.Printf("Warning: %s is already defined. Overriding it.\n", rule)
		}

		p.Rules[rule] = &PluginRule{Disable: conf.Disable, File: conf.File, Options: conf.Options}
	}

	return nil
}

func (p *PluginManager) OverrideRules(userOverrides map[string]PluginUserOverride) error {
	for rule, conf := range userOverrides {
		if val, ok := p.Rules[rule]; ok {
			if conf.Disable != nil {
				val.Disable = *conf.Disable
			}
			if conf.Options != nil {
				for i, r := range conf.Options {
					if val.Options == nil {
						val.Options = make(map[string]any, 0)
					}
					val.Options[i] = r
				}
			}
			p.Rules[rule] = val
		} else {
			fmt.Printf("Overriding rule %s not found\n", rule)
		}
	}

	return nil
}

func (p *PluginManager) ReadPluginCode(path string) (string, error) {
	data, err := p.Reader.ReadIntoRawBytes(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
