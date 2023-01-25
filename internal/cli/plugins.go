package cli

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

func (c *Compiler) Run(pgm *goja.Program) error {
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
	call(goja.Undefined())
	return nil
}

// https://github.com/grafana/k6/blob/a2a5b39017e371f4d72b85d4d10cbc1ac449f1ff/js/bundle.go#L328
// func StartRuntime() (*Runtime, error) {
// 	reqistry := new(require.Registry)

// 	runtime := goja.New()
// 	reqistry.Enable(runtime)
// 	console.Enable(runtime)

// 	babelProgram, err := goja.Compile("babel.js", babelBundle, false)

// 	if err != nil {
// 		return nil, err
// 	}

// 	_, err = runtime.RunProgram(babelProgram)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var babelTransformer goja.Callable

// 	babel := runtime.Get("Babel")
// 	if err := runtime.ExportTo(babel.ToObject(runtime).Get("transform"), &babelTransformer); err != nil {
// 		return nil, err
// 	}

// 	v, err := babelTransformer(babel, runtime.ToValue(jsRawCode), runtime.ToValue(map[string]interface{}{
// 		"presets": []string{"env"},
// 	}))

// 	// if err != nil {
// 	// 	panic(err)
// 	// }

// 	// code := v.ToObject(runtime).Get("code").String()
// 	// if _, err := runtime.RunString(code); err != nil {
// 	// 	panic(err)
// 	// }
// 	return &Runtime{babelTransformer: func(code string) error {} },nil
// }

// Builtin will packaged as zip
// Installed on runtime
type Plugins struct{}
