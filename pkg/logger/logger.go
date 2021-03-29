package logger

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func Debug(s string) {
	pc, fileName, _, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()
	fmt.Printf("%s %s: %s\n", name, filepath.Base(fileName), s)
}
