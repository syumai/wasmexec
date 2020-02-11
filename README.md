# wasmexec

* wasmexec is a utility to execute WebAssembly in Go with interface similar to WebAssembly.instantiateStreaming().
* This package is using [wagon](https://github.com/go-interpreter/wagon) as an interpreter of WebAssembly.

## Usage

### 1. Create Wasm binary

* multiply.wat

```wat
(module
  (import "go" "imultiply" (func $imultiply (param $a i32) (param $b i32) (result i32)))
  (func (export "multiply") (param $a i32) (param $b i32) (result i32)
    local.get $a
    local.get $b
    call $imultiply)
)
```

```console
$ wat2wasm multiply.wat
```

### 2. Read Wasm binary and execute exported function

* main.go

```go
// 1. Open Wasm binary
f, err := os.Open("example.wasm")
if err != nil { 
    return nil, err
}
defer f.Close()

// 2. Create ImportObject
importObject := wasmexec.ImportObject{
    "go": {
        "imultiply": func(a int32, b int32) int32 {
            return a * b
        },
    },
}

// 3. Initialize WebAssembly instance
inst, err := wasmexec.InstantiateStreaming(f, importObject)
if err != nil {
    return nil, err
}

// 4. Call exported function
result, err := inst.Call("multiply", uint64(2), uint64(3))
if err != nil {
    return nil, err
}
fmt.Println(result) // #=> 6
```

For details, please see `examples/multiply`.

## Supported Features

* [x] import func
* [ ] import global
* [ ] import memory

## License

MIT

## Author

syumai
