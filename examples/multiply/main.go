package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/syumai/wasmexec"
)

func main() {
	os.Exit(realMain(os.Stdout, os.Args))
}

func realMain(w io.Writer, args []string) int {
	if len(args) < 3 {
		fmt.Printf("args must be given\n")
		return 1
	}

	var intArgs []int
	for _, s := range args[1:] {
		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintf(w, "args must be integer\n")
			return 1
		}
		intArgs = append(intArgs, i)
	}

	f, err := os.Open("multiply.wasm")
	if err != nil {
		fmt.Fprintf(w, "failed to open wasm file\n")
		return 1
	}
	defer f.Close()

	importObject := wasmexec.ImportObject{
		"go": {
			"imultiply": func(a int, b int) int {
				return a * b
			},
		},
	}
	inst, err := wasmexec.InstantiateStreaming(f, importObject)
	if err != nil {
		fmt.Fprintf(w, "unexpected error: %v\n", err)
		return 1
	}

	result, err := inst.Call("multiply", uint64(intArgs[0]), uint64(intArgs[1]))
	if err != nil {
		fmt.Fprintf(w, "unexpected error: %v\n", err)
		return 1
	}
	fmt.Fprintln(w, result)
	return 0
}
