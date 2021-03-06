package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	if len(args) == 0 {
		panic("condition failed")
	}
	format := args[0].(string)
	args = args[1:]
	if len(args) == 0 {
		panic(format)
	}
	panic(fmt.Sprintf(format, args...))
}

func readFileMust(path string) []byte {
	d, err := ioutil.ReadFile(path)
	must(err)
	return d
}

func normalizePath(s string) string {
	return strings.Replace(s, `\`, `/`, -1)
}

func normalizeNewlines(d []byte) []byte {
	// replace CR LF (windows) with LF (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)
	// replace CF (mac) with LF (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}

func toLines(d []byte) []string {
	d = normalizeNewlines(d)
	s := string(d)
	return strings.Split(s, "\n")
}

func readFileAsLines(path string) ([]string, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return toLines(d), nil
}

func collapseMultipleEmptyLines(a []string) []string {
	j := 0
	prevWasEmpty := false
	for i := 0; i < len(a); i++ {
		l := a[i]
		isEmpty := len(l) == 0
		// skip n-th empty line in a row
		if isEmpty && prevWasEmpty {
			continue
		}
		prevWasEmpty = isEmpty
		a[j] = l
		j++
	}
	if j < len(a) {
		return a[:j]
	}
	return a
}

func dumpFile(path string) {
	fmt.Printf("File: %s\n", path)
	d, err := ioutil.ReadFile(path)
	must(err)
	fmt.Printf("%s\n", string(d))
}

func shortAPIName(fn *FunctionInfo) string {
	sfn := fn.SourceFile.FileName
	modName := fn.Module.Name
	fnName := fn.Function.Name
	return fmt.Sprintf(`%s %s.%s`, sfn, modName, fnName)
}

func gofmtFile(path string) {
	cmd := exec.Command("gofmt", "-s", "-w", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("gofmt failed with %s. Output:\n%s\n", err, string(out))
	}
}

func makeNameGoPublic(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
