package utils

import (
	"os"
	"sync"
)

type SafeFileWriter struct {
	file *os.File
	mu   sync.Mutex
}

func (sw *SafeFileWriter) write(data string) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.file.WriteString(data)
}
