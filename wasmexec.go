package wasmexec

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
)

type ImportObject map[string]map[string]interface{}

var kindOfInt reflect.Kind

func init() {
	sizeofInt := unsafe.Sizeof(int(0))
	switch sizeofInt {
	case 4:
		kindOfInt = reflect.Int32
	case 8:
		kindOfInt = reflect.Int64
	default:
		panic(fmt.Errorf("unexpected size of int: %d", sizeofInt))
	}
}

func toValueTypeFromKind(k reflect.Kind) (wasm.ValueType, error) {
	if k == reflect.Int {
		k = kindOfInt
	}
	switch k {
	case reflect.Int32, reflect.Uint32:
		return wasm.ValueTypeI32, nil
	case reflect.Int64, reflect.Uint64:
		return wasm.ValueTypeI64, nil
	case reflect.Float32:
		return wasm.ValueTypeF32, nil
	case reflect.Float64:
		return wasm.ValueTypeF64, nil
	}
	return 0, fmt.Errorf("unsupported kind: %q", k.String())
}

func generateResolverFromImportObject(importObject ImportObject) wasm.ResolveFunc {
	return func(name string) (*wasm.Module, error) {
		obj, ok := importObject[name]
		if !ok {
			return nil, fmt.Errorf("module %q was not found in multiply object", name)
		}
		m := wasm.NewModule()
		m.Export.Entries = make(map[string]wasm.ExportEntry)
		var index uint32
		for k, v := range obj {
			t := reflect.TypeOf(v)
			if t.Kind() != reflect.Func {
				return nil, fmt.Errorf("kind %q is not supported for multiply object", t.Kind())
			}

			var (
				ins, outs []reflect.Type
			)
			ins = append(ins, reflect.TypeOf(&exec.Process{}))

			// Create FunctionSignature
			sig := wasm.FunctionSig{Form: 0}
			for i := 0; i < t.NumIn(); i++ {
				p := t.In(i)
				ins = append(ins, p)
				vt, err := toValueTypeFromKind(p.Kind())
				if err != nil {
					fmt.Errorf("failed to convert kind of param %q: %w", p.Name(), err)
				}
				sig.ParamTypes = append(sig.ParamTypes, vt)
			}
			for i := 0; i < t.NumOut(); i++ {
				p := t.Out(i)
				outs = append(outs, p)
				vt, err := toValueTypeFromKind(p.Kind())
				if err != nil {
					fmt.Errorf("failed to convert kind of return value %q: %w", p.Name(), err)
				}
				sig.ReturnTypes = append(sig.ReturnTypes, vt)
			}
			m.Types.Entries = append(m.Types.Entries, sig)

			fnType := reflect.FuncOf(ins, outs, t.IsVariadic())
			fnValue := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
				origFunc := reflect.ValueOf(v)
				return origFunc.Call(args[1:]) // ignore *exec.Process
			})

			// Create FunctionIndexSpace
			indexSpace := wasm.Function{
				Sig:  &sig,
				Host: fnValue,
				Body: &wasm.FunctionBody{}, // dummy wasm body
			}
			m.FunctionIndexSpace = append(m.FunctionIndexSpace, indexSpace)

			// Add Function to Export
			m.Export.Entries[k] = wasm.ExportEntry{
				FieldStr: k,
				Kind:     wasm.ExternalFunction,
				Index:    index,
			}
			index++
		}
		return m, nil
	}
}

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

type Instance struct {
	Module *wasm.Module
	VM     *exec.VM
}

func (i *Instance) Call(name string, args ...uint64) (uint32, error) {
	fn, ok := i.Module.Export.Entries[name]
	if !ok {
		return 0, fmt.Errorf("failed to get func %q from wasm module", name)
	}
	o, err := i.VM.ExecCode(int64(fn.Index), args...)
	if err != nil {
		return 0, fmt.Errorf("unexpected error occured on execute func %q: %w", name, err)
	}
	result, ok := o.(uint32)
	if !ok {
		return 0, fmt.Errorf("return type must be uint32")
	}
	return result, nil
}
