package utils

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type SafeFileWriter struct {
	File *os.File
	Mu   sync.Mutex
}

func (sw *SafeFileWriter) Write(data string) (int, error) {
	sw.Mu.Lock()
	defer sw.Mu.Unlock()
	return sw.File.WriteString(data)
}

func GetGoroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, _ := strconv.Atoi(idField)
	return id
}
