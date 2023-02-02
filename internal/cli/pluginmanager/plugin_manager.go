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
	Rules  map[string]struct{ File string }
	Reader Reader
}

type PluginConfig struct {
	Rules []PluginConfigRule `yaml:"rules" `
}

type PluginConfigRule struct {
	Name string `yaml:"name"`
	File string `yam:"file"`
}

func New(fr Reader) *PluginManager {
	return &PluginManager{
		Rules:  map[string]struct{ File string }{},
		Reader: fr,
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
	// TODO(akhilmhdh): In prod mode this should point to "/.apic/plugins"
	path := filepath.Clean(filepath.Join(cwd, "./plugins/builtin"))
	pluginsDir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal("Failed to open builtin plugins dir: ", err)
	}

	for _, pluginDir := range pluginsDir {
		if pluginDir.IsDir() {
			pluginName := pluginDir.Name()
			pluginFolder := filepath.Join(path, fmt.Sprintf("/%s", pluginName))
			pluginFiles, err := os.ReadDir(pluginFolder)
			if err != nil {
				return err
			}
			// get plugin config file.
			pluginCfgName, err := getPluginConfFile(pluginFiles)
			if err != nil {
				return err
			}

			// load plugin config
			cfgFilePath := filepath.Join(path, fmt.Sprintf("/%s/%s", pluginName, pluginCfgName))
			var pluginCfg PluginConfig
			if err := p.Reader.ReadFile(cfgFilePath, &pluginCfg); err != nil {
				return err
			}

			// load up the rules
			for _, r := range pluginCfg.Rules {
				jsRuleFile := filepath.Join(pluginFolder, fmt.Sprintf("/%s", r.File))
				p.Rules[r.Name] = struct{ File string }{File: jsRuleFile}
			}
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
