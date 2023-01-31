// This compiler was inspired from grafana k6 project
// A big kudos goes to goja library
// Do check it out: https://github.com/dop251/goja
package compiler

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/dop251/goja"
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

func New() (*Compiler, error) {
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
	// load up  and compile babel file only once in a run
	// for concurrency
	compileBabelOnce.Do(func() {
		globalBabelCode, errGlobalBabelCode = goja.Compile("babel.js", babelBundle, false)
	})

	if errGlobalBabelCode != nil {
		return nil, errGlobalBabelCode
	}

	// JS console statement support
	// TODO(akhilmhdh): change this to centralized logger one
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

// these are the data that will be given to js code execution env
type RunConfig struct {
	ApiSchema map[string]interface{} `json:"schema"`
	Type      string                 `json:"type"`
}

func (c *Compiler) Run(pgm *goja.Program, cfg *RunConfig) error {
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
	call(goja.Undefined(), c.babel.runtime.ToValue(cfg))

	return nil
}
