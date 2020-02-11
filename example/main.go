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

func bytesToInt32(b [4]byte) int32 {
	var result int32
	for i := 0; i < 4; i++ {
		result += int32(b[i]) * int32(1<<(i*8))
	}
	return result
}

func realMain(w io.Writer, args []string) int {
	if len(args) < 4 {
		fmt.Printf("args must be given\n")
		return 1
	}

	var intArgs []int
	for _, s := range args[2:] {
		i, err := strconv.Atoi(s)
		if err != nil {
			fmt.Fprintf(w, "args must be integer\n")
			return 1
		}
		intArgs = append(intArgs, i)
	}

	f, err := os.Open("example.wasm")
	if err != nil {
		fmt.Fprintf(w, "failed to open wasm file\n")
		return 1
	}
	defer f.Close()

	//mem := wasmexec.NewMemory(1, 0) // WIP
	importObject := wasmexec.ImportObject{
		"go": {
			"imultiply": func(a int, b int) int {
				return a * b
			},
			//"memory": mem, // WIP
		},
	}

	inst, err := wasmexec.InstantiateStreaming(f, importObject)
	if err != nil {
		fmt.Fprintf(w, "unexpected error: %v\n", err)
		return 1
	}

	switch args[1] {
	case "add":
		_, err := inst.Call("add", uint64(intArgs[0]), uint64(intArgs[1]))
		if err != nil {
			fmt.Fprintf(w, "unexpected error: %v\n", err)
			return 1
		}

		aBytes := inst.VM.Memory()[0:4]
		bBytes := inst.VM.Memory()[4:8]
		a := bytesToInt32([4]byte{aBytes[0], aBytes[1], aBytes[2], aBytes[3]})
		b := bytesToInt32([4]byte{bBytes[0], bBytes[1], bBytes[2], bBytes[3]})
		fmt.Fprintln(w, a+b)
		return 0
	case "multiply":
		result, err := inst.Call("multiply", uint64(intArgs[0]), uint64(intArgs[1]))
		if err != nil {
			fmt.Fprintf(w, "unexpected error: %v\n", err)
			return 1
		}
		fmt.Fprintln(w, result)
		return 0
	}
	fmt.Fprintf(w, "command %q was not found", args[1])
	return 1
}
