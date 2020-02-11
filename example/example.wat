(module
  (import "go" "imultiply" (func $imultiply (param $a i32) (param $b i32) (result i32)))
  (memory 1)
  (func (export "multiply") (param $a i32) (param $b i32) (result i32)
    local.get $a
    local.get $b
    call $imultiply)
  (func (export "add") (param $a i32) (param $b i32) (result i32)
    i32.const 0
    local.get $a
    i32.store
    i32.const 4
    local.get $b
    i32.store
    i32.const 0)
)
