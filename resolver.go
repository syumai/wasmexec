package wasmexec

import (
	"fmt"
	"reflect"

	"github.com/go-interpreter/wagon/exec"
	"github.com/go-interpreter/wagon/wasm"
)

func generateResolverFromImportObject(importObject ImportObject) wasm.ResolveFunc {
	return func(name string) (*wasm.Module, error) {
		obj, ok := importObject[name]
		if !ok {
			return nil, fmt.Errorf("module %q was not found in multiply object", name)
		}
		m := wasm.NewModule()
		m.Export.Entries = make(map[string]wasm.ExportEntry)
		var memoryIndex, funcIndex uint32
		for k, v := range obj {
			rt := reflect.TypeOf(v)
			switch mem := v.(type) {
			case *Memory:
				panic("WIP")
				sectionMemories := &wasm.SectionMemories{}
				memoryEntry := wasm.Memory{
					Limits: wasm.ResizableLimits{
						Initial: mem.initial,
					},
				}
				if mem.maximum > 0 {
					memoryEntry.Limits.Flags = 1
					memoryEntry.Limits.Maximum = mem.maximum
				}
				sectionMemories.Entries = append(sectionMemories.Entries, memoryEntry)
				m.Memory = sectionMemories
				m.LinearMemoryIndexSpace = make([][]byte, 1)
				m.LinearMemoryIndexSpace[0] = mem.Buffer()
				exportEntry := wasm.ExportEntry{
					FieldStr: k,
					Kind:     wasm.ExternalMemory,
					Index:    memoryIndex,
				}
				m.Export.Entries[k] = exportEntry
				memoryIndex++
				continue
			}
			switch rt.Kind() {
			case reflect.Func:
				sig, indexSpace, exportEntry := createFunctionSignature(k, v, funcIndex)
				m.Types.Entries = append(m.Types.Entries, sig)
				m.FunctionIndexSpace = append(m.FunctionIndexSpace, indexSpace)
				m.Export.Entries[k] = exportEntry
				funcIndex++
				continue
			default:
				return nil, fmt.Errorf("kind %q is not supported for multiply object", rt.Kind())
			}
		}

		return m, nil
	}
}

func createFunctionSignature(name string, fn interface{}, index uint32) (wasm.FunctionSig, wasm.Function, wasm.ExportEntry) {
	t := reflect.TypeOf(fn)

	var (
		ins, outs               []reflect.Type
		paramTypes, returnTypes []wasm.ValueType
	)

	// Add *exec.Process as first argument (wagon requires this)
	ins = append(ins, reflect.TypeOf(&exec.Process{}))

	// Create ins, outs / paramTypes, returnTypes
	for i := 0; i < t.NumIn(); i++ {
		p := t.In(i)
		if p.Kind() == reflect.Int {
			// Always treat int as int32
			p = reflect.TypeOf(int32(0))
		}
		ins = append(ins, p)
		vt, err := toValueTypeFromKind(p.Kind())
		if err != nil {
			fmt.Errorf("failed to convert kind of param %q: %w", p.Name(), err)
		}
		paramTypes = append(paramTypes, vt)
	}
	for i := 0; i < t.NumOut(); i++ {
		p := t.Out(i)
		if p.Kind() == reflect.Int {
			// Always treat int as int32
			p = reflect.TypeOf(int32(0))
		}
		outs = append(outs, p)
		vt, err := toValueTypeFromKind(p.Kind())
		if err != nil {
			fmt.Errorf("failed to convert kind of return value %q: %w", p.Name(), err)
		}
		returnTypes = append(returnTypes, vt)
	}

	// Create wagon compatible func signature
	fnType := reflect.FuncOf(ins, outs, t.IsVariadic())
	fnValue := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		var realArgs []reflect.Value
		for i := 1; i < len(args); i++ { // ignore *exec.Process
			p := t.In(i - 1)
			if p.Kind() == reflect.Int {
				// int's internal interface is int (args is given as int32)
				realArgs = append(realArgs, reflect.ValueOf(args[i].Interface()).Convert(reflect.TypeOf(0)))
				continue
			}
			realArgs = append(realArgs, args[i])
		}
		origFunc := reflect.ValueOf(fn)
		returnValues := origFunc.Call(realArgs)

		var realReturnValues []reflect.Value
		for i := 0; i < len(returnValues); i++ {
			p := t.Out(i)
			if p.Kind() == reflect.Int {
				// int's external interface is int32
				realReturnValues = append(realReturnValues, reflect.ValueOf(returnValues[i].Interface()).Convert(reflect.TypeOf(int32(0))))
				continue
			}
			realReturnValues = append(realReturnValues, returnValues[i])
		}
		return realReturnValues
	})

	// Create FunctionSignature
	sig := wasm.FunctionSig{
		Form:        0,
		ParamTypes:  paramTypes,
		ReturnTypes: returnTypes,
	}

	// Create FunctionIndexSpace
	indexSpace := wasm.Function{
		Sig:  &sig,
		Host: fnValue,
		Body: &wasm.FunctionBody{}, // dummy wasm body
	}

	// Create ExportEntry
	exportEntry := wasm.ExportEntry{
		FieldStr: name,
		Kind:     wasm.ExternalFunction,
		Index:    index,
	}

	return sig, indexSpace, exportEntry
}

func toValueTypeFromKind(k reflect.Kind) (wasm.ValueType, error) {
	switch k {
	case reflect.Int:
		// Always treat int as int32.
		return wasm.ValueTypeI32, nil
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
