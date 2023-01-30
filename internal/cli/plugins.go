package cli

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/dop251/goja"
	"gopkg.in/yaml.v3"
)

//go:embed babel.min.js
var babelBundle string

var (
	compileBabelOnce   sync.Once
	globalBabelCode    *goja.Program
	errGlobalBabelCode error
)

type Compiler struct {
	babel *Babel
}

func NewCompiler() (*Compiler, error) {
	b, err := newBabel()
	if err != nil {
		return nil, err
	}

	return &Compiler{babel: b}, nil
}

type Babel struct {
	runtime     *goja.Runtime
	this        goja.Value
	transformer goja.Callable
}

// console logger
func newBabel() (*Babel, error) {
	compileBabelOnce.Do(func() {
		globalBabelCode, errGlobalBabelCode = goja.Compile("babel.js", babelBundle, false)
	})

	if errGlobalBabelCode != nil {
		return nil, errGlobalBabelCode
	}
	logger := func(fc goja.FunctionCall) goja.Value {
		fmt.Println(fc.Arguments)
		return nil
	}

	runtime := goja.New()
	runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	runtime.Set("console", map[string]func(goja.FunctionCall) goja.Value{
		"log":   logger,
		"error": logger,
		"warn":  logger,
	})

	// runtime.Set("exports", map[string]interface{}{})

	_, err := runtime.RunProgram(globalBabelCode)
	if err != nil {
		return nil, err
	}

	babel := runtime.Get("Babel")
	b := &Babel{runtime: runtime, this: babel}

	if err := runtime.ExportTo(babel.ToObject(runtime).Get("transform"), &b.transformer); err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Compiler) Transform(rawCode string) (*goja.Program, error) {

	v, err := c.babel.transformer(c.babel.this, c.babel.runtime.ToValue(rawCode), c.babel.runtime.ToValue(map[string]interface{}{
		"presets": []string{"env"},
	}))

	if err != nil {
		return nil, err
	}

	code := v.ToObject(c.babel.runtime).Get("code").String()
	pgm, err := goja.Compile("test", fmt.Sprintf(`(function(exports){
		%s
		})`, code), true)

	if err != nil {
		return nil, err
	}

	return pgm, nil
}

type RunConfig struct {
	Data map[string]interface{} `json:"data"`
	Type string                 `json:"type"`
}

func (c *Compiler) Run(pgm *goja.Program, cfg *RunConfig) error {
	v, err := c.babel.runtime.RunProgram(pgm)
	if err != nil {
		return err
	}
	call, ok := goja.AssertFunction(v)
	if !ok {
		return fmt.Errorf("failed to get exports")
	}

	export := c.babel.runtime.NewObject()
	// runtime.Set("exports", map[string]interface{}{})
	call(goja.Undefined(), export)

	fn := export.Get("default")
	call, ok = goja.AssertFunction(fn)
	if !ok {
		return fmt.Errorf("failed to get exports")
	}
	call(goja.Undefined(), c.babel.runtime.ToValue(cfg))
	return nil
}

// Builtin will packaged as zip
// Installed on runtime
type PluginManager struct {
	rules map[string]struct{ File string }
}

type PluginConfig struct {
	Rules []PluginConfigRule `yaml:"rules"`
}

type PluginConfigRule struct {
	Name string `yaml:"name"`
	File string `yam:"file"`
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		rules: map[string]struct{ File string }{},
	}
}

func (p *PluginManager) LoadBuiltinPlugin() error {
	// this will soon be a root config dir specially for apic plugins
	files, err := os.ReadDir("./internal/builtin")
	if err != nil {
		log.Fatal("Failed to open builtin plugins dir: ", err)
	}

	for _, file := range files {
		if file.IsDir() {
			data, err := os.ReadFile(fmt.Sprintf("./internal/builtin/%s/config.yaml", file.Name()))
			if err != nil {
				return fmt.Errorf("failed to open builtin plugin file. Config file not found %s", file.Name())
			}

			// load plugin config
			var pluginCfg PluginConfig
			if err = yaml.Unmarshal(data, &pluginCfg); err != nil {
				return err
			}

			// load up the rules
			for _, r := range pluginCfg.Rules {
				filePath := fmt.Sprintf("./internal/builtin/%s/%s", file.Name(), r.File)
				p.rules[r.Name] = struct{ File string }{File: filePath}
			}
		}
	}

	return nil
}

func (p *PluginManager) ReadPluginCode(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
