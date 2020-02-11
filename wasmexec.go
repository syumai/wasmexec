package wasmexec

import (
	"fmt"
	"io"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
)

type ImportObject map[string]map[string]interface{}

func InstantiateStreaming(src io.Reader, importObject ImportObject) (*Instance, error) {
	module, err := wasm.ReadModule(src, generateResolverFromImportObject(importObject))
	if err != nil {
		return nil, fmt.Errorf("unexpected error occured on reading module: %v", err)
	}
	vm, err := exec.NewVM(module)
	if err != nil {
		return nil, fmt.Errorf("unexpected error occured on creating VM: %v", err)
	}
	return &Instance{
		Module: module,
		VM:     vm,
	}, nil
}
