(module
  (import "go" "imultiply" (func $imultiply (param $a i32) (param $b i32) (result i32)))
  (func (export "multiply") (param $a i32) (param $b i32) (result i32)
    local.get $a
    local.get $b
    call $imultiply)
)
