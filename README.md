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
f, _ := os.Open("exmaple.wasm")
defer f.Close()

// 2. Create ImportObject
importObject := wasmexec.ImportObject{
    "go": {
        "imultiply": func(a int, b int) int{
            return a * b
        },
    },
}

// 3. Initialize WebAssembly instance
inst, _ := wasmexec.InstantiateStreaming(f, importObject)

// 4. Call exported function
result, _ := inst.Call("multiply", uint64(2), uint64(3))
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
