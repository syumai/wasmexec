package wasmexec

import (
	"fmt"
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
)

type Instance struct {
	Module *wasm.Module
	VM     *exec.VM
}

func (i *Instance) Call(name string, args ...uint64) (uint64, error) {
	fn, ok := i.Module.Export.Entries[name]
	if !ok {
		return 0, fmt.Errorf("failed to get func %q from wasm module", name)
	}
	o, err := i.VM.ExecCode(int64(fn.Index), args...)
	if err != nil {
		return 0, fmt.Errorf("unexpected error occured on execute func %q: %w", name, err)
	}
	result, ok := reflect.ValueOf(o).Convert(reflect.TypeOf(uint64(0))).Interface().(uint64)
	if !ok {
		return 0, fmt.Errorf("return type must be converted to uint64")
	}
	return result, nil
}
