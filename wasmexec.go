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

	// WIP
	// This fixes buggy move in wagon.
	// resolveImports doesn't set Memory.Entries[0].Initial
	// https://github.com/go-interpreter/wagon/blob/8dd99d5adacc2fcb2a87b96c6001ecc84a6ed09c/wasm/imports.go#L215
	sectionMemories := &wasm.SectionMemories{}
	memoryEntry := wasm.Memory{
		Limits: wasm.ResizableLimits{
			Initial: 1, // This value must be given from importObject
		},
	}
	sectionMemories.Entries = append(sectionMemories.Entries, memoryEntry)
	module.Memory = sectionMemories

	vm, err := exec.NewVM(module)
	if err != nil {
		return nil, fmt.Errorf("unexpected error occured on creating VM: %v", err)
	}
	return &Instance{
		Module: module,
		VM:     vm,
	}, nil
}
