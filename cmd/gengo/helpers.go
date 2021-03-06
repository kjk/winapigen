package main

import (
	"fmt"
)

func syscallFuncForArgCount(n int) (string, int) {
	if n <= 3 {
		return "syscall.Syscall", 3
	}
	if n <= 6 {
		return "syscall.Syscall6", 6
	}
	if n <= 9 {
		return "syscall.Syscall9", 9
	}
	if n <= 12 {
		return "syscall.Syscall12", 12
	}
	if n <= 15 {
		return "syscall.Syscall15", 15
	}
	s := fmt.Sprintf("unsupported number of arguments: %d", n)
	panic(s)
}
