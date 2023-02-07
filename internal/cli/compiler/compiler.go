// This compiler was inspired from grafana k6 project
// A big kudos goes to goja library
// Do check it out: https://github.com/dop251/goja
package compiler

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/1-platform/api-catalog/internal/cli/compiler/modules"
	"github.com/1-platform/api-catalog/internal/cli/reportmanager"
	"github.com/dop251/goja"
)

//go:embed babel.min.js
var babelBundle string

var (
	compileBabelOnce   sync.Once
	globalBabelCode    *goja.Program
	errGlobalBabelCode error
)

var ErrExceptionInPluginCode = errors.New("plugin code error")

type Logger interface {
	Info(str string)
	Warn(str string)
	Error(str string)
	Log(str string)
}

type Compiler struct {
	babel        *Babel
	ModuleLoader *modules.ModuleLoader
	logger       Logger
}

func New(logger Logger) (*Compiler, error) {
	runtime := NewRuntime()
	setConsole(runtime, logger)
	moduleLoader := modules.New(runtime)

	b, err := newBabel(runtime)
	if err != nil {
		return nil, err
	}
	cmp := &Compiler{babel: b, ModuleLoader: moduleLoader, logger: logger}

	return cmp, nil
}

type Babel struct {
	runtime     *goja.Runtime
	this        goja.Value
	transformer goja.Callable
}

func NewRuntime() *goja.Runtime {
	runtime := goja.New()
	runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	return runtime
}

// to support console logging
func setConsole(runtime *goja.Runtime, logger Logger) {
	runtime.Set("console", map[string]func(goja.FunctionCall) goja.Value{
		"log": func(fc goja.FunctionCall) goja.Value {
			var sb strings.Builder
			for _, r := range fc.Arguments {
				sb.WriteString(r.String())
			}
			logger.Log(sb.String())
			return nil
		},
		"error": func(fc goja.FunctionCall) goja.Value {
			var sb strings.Builder
			for _, r := range fc.Arguments {
				sb.WriteString(r.String())
			}
			logger.Error(sb.String())
			return nil
		},
		"warn": func(fc goja.FunctionCall) goja.Value {
			var sb strings.Builder
			for _, r := range fc.Arguments {
				sb.WriteString(r.String())
			}
			logger.Warn(sb.String())
			return nil
		},
	})

}

// console logger
func newBabel(runtime *goja.Runtime) (*Babel, error) {
	// load up  and compile babel file only once in a run
	// for concurrency
	compileBabelOnce.Do(func() {
		globalBabelCode, errGlobalBabelCode = goja.Compile("babel.js", babelBundle, false)
	})

	if errGlobalBabelCode != nil {
		return nil, errGlobalBabelCode
	}

	_, err := runtime.RunProgram(globalBabelCode)
	if err != nil {
		return nil, err
	}

	// REF: https://babeljs.io/docs/en/babel-standalone
	babel := runtime.Get("Babel")
	b := &Babel{runtime: runtime, this: babel}

	if err := runtime.ExportTo(babel.ToObject(runtime).Get("transform"), &b.transformer); err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Compiler) Transform(rawCode string) (*goja.Program, error) {
	// change the code to commonjs using babel
	v, err := c.babel.transformer(c.babel.this, c.babel.runtime.ToValue(rawCode), c.babel.runtime.ToValue(map[string]interface{}{
		"presets": []string{"env"},
	}))
	if err != nil {
		return nil, err
	}

	code := v.ToObject(c.babel.runtime).Get("code").String()
	// wrap the commonjs module inside a function
	// This will private scope each functions we execute
	// Compile to a goja program thus can be executed anytime with goja
	pgm, err := goja.Compile("test", fmt.Sprintf(`(function(exports){
		%s
		})`, code), true)

	if err != nil {
		return nil, err
	}

	return pgm, nil
}

type KeyValuePairs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// these are the data that will be given to js code execution env
type RunConfig struct {
	ApiSchema map[string]interface{}               `json:"schema"`
	Type      string                               `json:"type"`
	SetScore  func(category string, score float32) `json:"setScore"`
	Report    func(body *reportmanager.ReportDef)  `json:"report"`
}

func (c *Compiler) Run(pgm *goja.Program, cfg *RunConfig, ruleOpt map[string]any) error {
	v, err := c.babel.runtime.RunProgram(pgm)
	if err != nil {
		return err
	}

	// the wrapped function is preparing to execute
	call, ok := goja.AssertFunction(v)
	if !ok {
		return fmt.Errorf("failed to get exports")
	}

	// the argument export
	export := c.babel.runtime.NewObject()
	// execute the wrapper function now export contains default function
	call(goja.Undefined(), export)

	// execute the default function with configuration passed
	fn := export.Get("default")
	call, ok = goja.AssertFunction(fn)
	if !ok {
		return fmt.Errorf("failed to get exports")
	}
	_, err = call(goja.Undefined(), c.babel.runtime.ToValue(cfg), c.babel.runtime.ToValue(ruleOpt))
	if err != nil {
		return fmt.Errorf("%s%w", err.Error(), ErrExceptionInPluginCode)
	}

	return nil
}
