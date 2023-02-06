package modules

import (
	"os"

	"github.com/dop251/goja"
)

// to get and set os environment values
// TODO(akhilmhdh): Add logger to know env manipulation
func (m *ModuleLoader) envModule() goja.Value {
	obj := m.runtime.NewObject()

	obj.Set("__esModule", true)
	obj.Set("setEnv", func(envName string, envValue string) {
		os.Setenv(envName, envValue)
	})
	obj.Set("getEnv", func(envName string) string {
		return os.Getenv(envName)
	})

	return obj
}
