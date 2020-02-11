package wasmexec

type Memory struct {
	initial uint32
	maximum uint32
	buffer  []byte
}

const MemoryPageSize = 65536 // 64KiB

func NewMemory(initial uint32, maximum uint32) *Memory {
	buf := make([]byte, MemoryPageSize)
	return &Memory{
		initial: initial,
		maximum: maximum,
		buffer:  buf,
	}
}

func (m *Memory) Buffer() []byte {
	return m.buffer
}
