package modules

import (
	"bytes"
	"os/exec"

	"github.com/dop251/goja"
)

type ExecCommandRun struct {
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Module to execute system commands
func (m *ModuleLoader) execCommandModule() goja.Value {
	obj := m.runtime.NewObject()
	// to allow default export
	obj.Set("__esModule", true)
	obj.Set("default", func(command string) *ExecCommandRun {
		cmd := exec.Command(command)
		var outb, errb bytes.Buffer

		cmd.Stdout = &outb
		cmd.Stderr = &errb

		if err := cmd.Run(); err != nil {
			return &ExecCommandRun{Data: outb.String(), Error: err.Error()}
		}

		return &ExecCommandRun{Data: outb.String(), Error: errb.String()}
	})

	return obj
}
