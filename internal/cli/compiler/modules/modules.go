package modules

import (
	"errors"
	"strings"

	"github.com/dop251/goja"
)

var errUnknownModule = errors.New("module not found")

type ModuleLoader struct {
	runtime        *goja.Runtime
	builtInModules map[string]goja.Value
}

func New(runtime *goja.Runtime) *ModuleLoader {
	mod := &ModuleLoader{
		runtime:        runtime,
		builtInModules: make(map[string]goja.Value),
	}
	mod.loadBuiltinModules()

	runtime.Set("require", mod.resolver)
	return mod
}

func (m *ModuleLoader) resolver(module string) goja.Value {
	if strings.HasPrefix(module, "apic/") {
		// load builtIn modules
		if val, ok := m.builtInModules[module]; ok {
			return val
		} else {
			panic(m.runtime.NewGoError(errUnknownModule))
		}
	}

	panic(m.runtime.NewGoError(errUnknownModule))
}

func (m *ModuleLoader) loadBuiltinModules() {
	m.builtInModules["apic/exec"] = m.execCommandModule()
	m.builtInModules["apic/env"] = m.envModule()
	m.builtInModules["apic/strings"] = m.stringModule()
}
