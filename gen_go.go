package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kylelemons/godebug/pretty"
)

const (

	// this is for Variable that don't belong to any module
	noModule = "nomodule.go"

	fileHdr = `package w

import (
	"golang.org/x/sys/windows"
	"syscall"
	"unsafe"
)
`
)

var (
	printGeneratedFile = false
)

// information needed to generate info for a single source
// file (which can map to .dll file or a .h include file)
type goSourceFile struct {
	// name of the .go file where we'll write generated info
	sourceFileName string

	functionsToGenerate []*FunctionInfo
	typesToGenerate     []*TypeInfo

	// interfaces don't come from DLLs but we reuse the
	// concept
	interfacesToGenerate []*InterfaceInfo
}

func (sf *goSourceFile) getModule() *Module {
	// can be nil for interfaces
	funcs := sf.functionsToGenerate
	if len(funcs) > 0 {
		mod := funcs[0].Module
		panicIf(mod == nil, "module should not be nil")
		// verify all modules are the same
		for _, fn := range funcs[1:] {
			mod2 := fn.Module
			panicIf(mod != mod2, "mod != mod2")
		}
		return mod
	}

	return nil
}

func (i *goSourceFile) Path() string {
	panicIf(i.sourceFileName == "")
	return filepath.Join("w", i.sourceFileName)
}

// goGenerator keeps info needed for generating Go files
// we generate one .go file per module
type goGenerator struct {
	// maps module (dll) name to info for generating .go
	// file for that module
	sourceFiles map[string]*goSourceFile

	// we keep track of which const values have already been generated
	// because they the same constant might belong to multiple enums / sets
	generatedConsts map[string]struct{}

	// the file we're currently writing to
	w io.Writer
}

func newGoGenerator() *goGenerator {
	return &goGenerator{
		sourceFiles:     map[string]*goSourceFile{},
		generatedConsts: map[string]struct{}{},
	}
}

func ws(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	must(err)
}

func (g *goGenerator) ws(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	ws(g.w, s)
}

func (g *goGenerator) rememberFunction(fi *FunctionInfo) *goSourceFile {
	fileName := fi.Module.goSourceFileName()
	si := g.getOrMakeSourceFileInfo(fileName)
	si.functionsToGenerate = append(si.functionsToGenerate, fi)
	return si
}

func (g *goGenerator) rememberType(vi *TypeInfo) *goSourceFile {
	fileName := vi.goSourceFileName()
	si := g.getOrMakeSourceFileInfo(fileName)
	si.typesToGenerate = append(si.typesToGenerate, vi)
	return si
}

func (g *goGenerator) rememberInterface(ii *InterfaceInfo) *goSourceFile {
	//fmt.Printf("remembering interface: %s\n", ii.Name())
	fileName := ii.goSourceFileName()
	si := g.getOrMakeSourceFileInfo(fileName)
	si.interfacesToGenerate = append(si.interfacesToGenerate, ii)
	return si
}

// maybe not the best way to do it
// replace e.g. **void with *uintptr
func fixGoTypeName(typeName string) string {
	if strings.HasSuffix(typeName, "*void") {
		//return strings.TrimSuffix(typeName, "*void") + "uintptr"
		return strings.TrimSuffix(typeName, "*void") + "unsafe.Pointer"
	}
	return typeName
}

// looks up a type by name. Returns a Go name of the type.
// when needed adds a type to typesToGenerate list. Not all types
// need to be generated (e.g. intermediary alias types)
func (g *goGenerator) addType(typeName string, vi *TypeInfo) string {
	// predefined types don't need to be generated
	if predefined := desugarPreDefinedType(typeName); predefined != "" {
		return predefined
	}

	if vi == nil {
		vi = findType(typeName)
	}

	if vi == nil {
		typeNoPointer := dePointerType(typeName)
		if typeNoPointer != "" {
			return "*" + g.addType(typeNoPointer, nil)
		}
	}

	if vi == nil {
		iname := g.addInterface(typeName)
		if iname != "" {
			return iname
		}
	}

	panicIf(vi == nil, "didn't find type with name '%s'", typeName)

	if vi.WasAdded {
		return vi.GoTypeName
	}
	vi.WasAdded = true

	v := vi.Variable
	if v.Type == typeVoid {
		vi.GoTypeName = "void"
		return vi.GoTypeName
	}

	if v.Type == typeModuleHandle {
		// this one is pre-defined
		vi.GoTypeName = "HANDLE"
		return vi.GoTypeName
	}

	if v.Type == typeUnion {
		fmt.Printf("union %s\n", v.Name)
		panic("union NYI")
	}

	if v.Type == typeStruct {
		panicIf(v.Flag != nil, "v.Flag on struct not nil")
		panicIf(v.Enum != nil, "v.Enum on struct not nil")
		panicIf(v.Set != nil, "v.Set on struct not nil")
		for _, t := range v.Field {
			typeName := g.addType(t.Type, nil)
			if typeName == "" {
				fmt.Printf("dindn't resolve type with t.Type = '%s'\n", t.Type)
				pretty.Print(t)
			}
			panicIf(typeName == "")
			typeName = fixGoTypeName(typeName)
			panicIf(t.Name == "")
			f := &NameAndType{
				Name:     t.Name,
				TypeName: typeName,
			}
			vi.Fields = append(vi.Fields, f)
		}
		g.rememberType(vi)
		vi.GoTypeName = v.Name
		return vi.GoTypeName
	}

	if v.Type == typePointer {
		// we don't add pointers
		vi.GoTypeName = "*" + g.addType(v.Base, nil)
		return vi.GoTypeName
	}

	if v.Type == typeArray {
		baseTypeName := g.addType(v.Base, nil)
		vi.GoTypeName = fmt.Sprintf("[%s]%s", v.Count, baseTypeName)
		return vi.GoTypeName
	}

	if v.Type == typeInteger {
		vi.GoTypeName = desugarInteger(v)
		return vi.GoTypeName
	}

	if v.Type == typeInterface {
		panic("interface NYI")
	}

	if v.Flag != nil {
		g.rememberType(vi)
		//fmt.Printf("Base for flag %s: %s\n", v.Name, v.Base)
		vi.GoTypeName = g.addType(v.Base, nil)
		return vi.GoTypeName
	}

	if v.Enum != nil {
		g.rememberType(vi)
		//fmt.Printf("Base for enum %s: %s\n", v.Name, v.Base)
		vi.GoTypeName = g.addType(v.Base, nil)
		return vi.GoTypeName
	}

	if v.Set != nil {
		panic("v.Set NYI")
		/*
			g.rememberType(vi)
			// TODO: desugar?
			fmt.Printf("Base for set %s: %s\n", v.Name, v.Base)
			vi.Name = v.Name
			return vi.Name
		*/
	}

	if v.Type == typeAlias {
		// we want to preserve types that are aliases for HANDLE
		// (HWND, HMENU etc.)
		if v.Base == "HANDLE" {
			vi.GoTypeName = v.Name
			g.rememberType(vi)
			return v.Name
		}
		vi.GoTypeName = g.addType(v.Base, nil)
		return vi.GoTypeName
	}

	panic(fmt.Sprintf("unknown type %s", v.Type))
}

func (g *goGenerator) getOrMakeSourceFileInfo(sourceFileName string) *goSourceFile {
	panicIf(sourceFileName == "", "name is empty")
	mi := g.sourceFiles[sourceFileName]
	if mi == nil {
		mi = &goSourceFile{
			sourceFileName: sourceFileName,
		}
		g.sourceFiles[sourceFileName] = mi
	}
	return mi
}

// types with name "IBindCtx*" is a pointer to an interface IBindCtx
// returns interface info or nil if s not a pointer to an interface
func (g *goGenerator) addInterface(name string) string {
	ii := allInterfaces[name]
	if ii == nil {
		fmt.Printf("didn't find interface named %s\n", name)
	}
	panicIf(ii == nil)
	if ii.WasAdded {
		return ii.Name()
	}
	ii.WasAdded = true

	g.rememberInterface(ii)
	i := ii.Definition.Interface

	for _, method := range i.API {
		api := method
		name := api.Name
		fi := &FunctionInfo{
			SourceFile: ii.Declaration.SourceFile,
			Function:   api,
			Name:       name,
		}

		g.buildFunctionInfo(fi)
		ii.Methods = append(ii.Methods, fi)
	}

	if i.BaseInterface != "" {
		baseName := g.addInterface(i.BaseInterface)
		panicIf(baseName == "", "didn't add interface '%s'", i.BaseInterface)
	}

	return ii.Name()
}

func (g *goGenerator) buildFunctionInfo(fi *FunctionInfo) {
	ret := fi.Function.Return
	goTypeName := g.addType(ret.Type, nil)
	panicIf(goTypeName == "")
	goTypeName = fixGoTypeName(goTypeName)
	fi.ReturnType = desugarReturnType(goTypeName)

	for _, arg := range fi.Function.Params {
		typeName := g.addType(arg.Type, nil)
		panicIf(typeName == "")
		typeName = fixGoTypeName(typeName)
		panicIf(arg.Name == "")
		farg := &NameAndType{
			Name:     arg.Name,
			TypeName: typeName,
		}
		fi.Args = append(fi.Args, farg)
	}
}

func (g *goGenerator) addFunction(name string) {
	fi := findFunction(name)
	panicIf(fi == nil, "didn't find function '%s'", name)
	if fi.WasAdded {
		return
	}
	fi.WasAdded = true

	g.rememberFunction(fi)
	g.buildFunctionInfo(fi)
}

func (g *goGenerator) generateConsts(set []*Set) bool {
	// TODO: optimize if only one value
	if len(set) == 0 {
		return false
	}
	g.ws("const (\n")
	for _, e := range set {
		name := e.Name
		// the same value can appear in multiple enum sets
		// so we need to ensure we only generate value once
		if _, ok := g.generatedConsts[name]; ok {
			continue
		}
		g.ws("%s = %s\n", name, e.Value)
		g.generatedConsts[name] = struct{}{}
	}
	g.ws(")\n\n")
	return true
}

func (g *goGenerator) generateType(ti *TypeInfo) {
	if ti.WasGenerated {
		return
	}
	ti.WasGenerated = true

	v := ti.Variable
	if v.Set != nil {
		g.generateConsts(v.Set)
		return
	}
	if v.Flag != nil {
		g.generateConsts(v.Flag.Set)
		return
	}
	if v.Enum != nil {
		g.generateConsts(v.Enum.Set)
		return
	}

	if v.Type == typeAlias {
		g.ws("type %s %s\n\n", ti.GoTypeName, v.Base)
		return
	}

	if v.Type == typeStruct {
		// first a pass to generate type names
		g.ws("type %s struct {\n", v.Name)
		for _, f := range ti.Fields {
			name := makeNameGoPublic(f.Name)
			g.ws("%s %s\n", name, f.TypeName)
		}
		g.ws("}\n\n")
		return
	}
}

func (g *goGenerator) generateFunctionBodies(mi *goSourceFile) {
	if len(mi.functionsToGenerate) == 0 {
		return
	}
	for _, fi := range mi.functionsToGenerate {
		g.generateFunction(fi)
	}
}

func (g *goGenerator) generateFunctionVariables(sf *goSourceFile) {
	if len(sf.functionsToGenerate) == 0 {
		return
	}
	mod := sf.getModule()
	panicIf(mod == nil, "mi.module should not be nil")
	modName := mod.moduleName()
	dllName := mod.dllName()

	g.ws("var (\n")

	// global variable that is a handle to dll
	g.ws("lib%s *windows.LazyDLL\n\n", modName)

	// for each function from dll that we use, a global variable
	// for the function address, like this:
	// abortDoc               *windows.LazyProc
	for _, fi := range sf.functionsToGenerate {
		g.ws("\t%s *windows.LazyProc\n", fi.GoVarName())
	}

	g.ws(")\n")
	/*
		// Generate:
		func init() {
			// Library
			libkernel32 = windows.NewLazySystemDLL("kernel32.dll")

			// Functions
			activateActCtx = libkernel32.NewProc("ActivateActCtx")
		}
	*/
	g.ws("\nfunc init() {\n")
	g.ws("lib%s = windows.NewLazySystemDLL(\"%s\")\n", modName, dllName)

	for _, fi := range sf.functionsToGenerate {
		g.ws("\t%s = lib%s.NewProc(\"%s\")\n", fi.GoVarName(), modName, fi.Name)
	}

	g.ws("}\n")
}

// if uses import (has a line like "windows.") then does nothing
// otherwise removes import lines (those that have e.g "golang.org/x/sys/windows")
func removeUnusedImportsInLines(lines []string, importUse, importPath string) []string {
	for _, line := range lines {
		if strings.Contains(line, importUse) {
			return lines
		}
	}
	n := 0
	nRemoved := 0
	for _, line := range lines {
		if strings.Contains(line, importPath) {
			nRemoved++
			continue
		}
		lines[n] = line
		n++
	}
	panicIf(nRemoved != 1)
	return lines[:len(lines)-1]
}

// this is a hacky way to do it, use goimports instead
func removeUnusedImportsInFile(path string) {
	imports := map[string]string{
		"windows.": `"golang.org/x/sys/windows"`,
		"syscall.": `"syscall"`,
		"unsafe.":  `"unsafe"`,
	}
	changed := false
	lines, err := readFileAsLines(path)
	must(err)
	for importUse, importPath := range imports {
		nBefore := len(lines)
		lines = removeUnusedImportsInLines(lines, importUse, importPath)
		changed = changed || (nBefore != len(lines))
	}
	if !changed {
		return
	}
	d := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path, []byte(d), 0644)
	must(err)
}

func (g *goGenerator) generateSourceFile(sf *goSourceFile) {
	//fmt.Printf("Generating module %s\n", sf.sourceFileName)
	var err error
	g.w, err = os.Create(sf.Path())
	must(err)
	defer func() {
		f := g.w.(*os.File)
		err := f.Close()
		must(err)
		g.w = nil
	}()

	g.ws(fileHdr)

	g.generateFunctionVariables(sf)

	for _, ti := range sf.typesToGenerate {
		g.generateType(ti)
	}

	g.generateFunctionBodies(sf)

	g.generateInterfaces(sf)
}

func desugarInteger(v *Variable) string {
	tp := "int"
	if v.Unsigned == "True" {
		tp = "u" + tp
	}
	switch v.Size {
	case "1":
		return tp + "8"
	case "2":
		return tp + "16"
	case "4":
		return tp + "32"
	case "8":
		return tp + "64"
	}

	//pretty.Print(v)
	panic("unsupported Variable of type Integer")
}

// if returns "", not a known type
func desugarPreDefinedType(tp string) string {
	// some known terminal types
	switch tp {
	case "HRESULT", "LARGE_INTEGER", "ULARGE_INTEGER", "WCHAR",
		"LPWSTR", "LPCWSTR":
		return tp
	}

	tp = strings.ToLower(tp)
	switch tp {
	case "lpvoid":
		return "*void"
	case "lpctstr":
		// TODO: this should depend on A vs. W
		return "*uint16"
	case "lpcwstr":
		return "*uint16"
	case "int8", "int16", "int32", "int64":
		return tp
	case "uint8", "uint16", "uint32", "uint64":
		return tp
	case "uint_ptr":
		return "uintptr"
	case "int":
		// yes, it's strange but that's what it seems to be
		return "int32"
	case "modulehandle":
		return "HANDLE"
	case "guid":
		return "GUID"
	}

	if strings.ToLower(tp) == "void" {
		return "void"
	}
	return ""
}

func desugarReturnType(typeName string) string {
	if typeName == "BOOL" {
		return "bool"
	}
	if strings.ToLower(typeName) == "void" {
		return ""
	}
	return typeName
}

/*
func CreateWindowEx(dwExStyle uint32, lpClassName, lpWindowName *uint16, dwStyle uint32, x, y, nWidth, nHeight int32, hWndParent HWND, hMenu HMENU, hInstance HINSTANCE, lpParam unsafe.Pointer) HWND {
	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
		uintptr(dwExStyle),
		uintptr(unsafe.Pointer(lpClassName)),
		uintptr(unsafe.Pointer(lpWindowName)),
		uintptr(dwStyle),
		uintptr(x),
		uintptr(y),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(hWndParent),
		uintptr(hMenu),
		uintptr(hInstance),
		uintptr(lpParam))

	return HWND(ret)
}
*/

func (g *goGenerator) generateFunction(fi *FunctionInfo) {
	fn := fi.Function

	returnType := fi.ReturnType
	s := ""
	for _, arg := range fi.Args {
		s += fmt.Sprintf("%s %s, ", arg.Name, arg.TypeName)
	}
	s = strings.TrimSuffix(s, ", ")
	g.ws("\nfunc %s(%s) %s {\n", fi.Name, s, returnType)

	// 	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
	nArgs := len(fn.Params)
	sysName, sysArgsCount := syscallFuncForArgCount(nArgs)
	if returnType != "" {
		g.ws("ret, _, _ := ")
	} else {
		g.ws("_, _, _ = ")
	}
	g.ws("%s(%s.Addr(), %d,\n", sysName, fi.GoVarName(), nArgs)
	for _, arg := range fi.Args {
		if dePointerType(arg.TypeName) != "" {
			g.ws("uintptr(unsafe.Pointer(%s)),\n", arg.Name)
		} else {
			g.ws("uintptr(%s),\n", arg.Name)
		}
	}
	nLeftOver := sysArgsCount - nArgs
	for nLeftOver > 0 {
		g.ws("0,\n")
		nLeftOver--
	}
	g.ws(")\n")
	if returnType != "" {
		if returnType == "bool" {
			g.ws("return ret != 0\n")
		} else {
			g.ws("return %s(ret)\n", returnType)
		}
	}

	g.ws("\n}\n")
}

func (g *goGenerator) generateInterfaces(mi *goSourceFile) {
	if len(mi.interfacesToGenerate) == 0 {
		return
	}
	for _, ii := range mi.interfacesToGenerate {
		g.generateInterface(ii)
	}
}

func splitGUID(s string) []string {
	s = strings.TrimPrefix(s, "{")
	s = strings.TrimSuffix(s, "}")

	parts := strings.Split(s, "-")
	panicIf(len(parts) != 5)
	a := []string{"0x" + parts[0]}
	a = append(a, "0x"+parts[1])
	a = append(a, "0x"+parts[2])
	s = parts[3]
	a = append(a, "0x"+s[:2])
	a = append(a, "0x"+s[2:])
	s = parts[4]
	for i := 0; i < 6; i++ {
		idx := i * 2
		s2 := "0x" + s[idx:idx+2]
		a = append(a, s2)
	}
	return a
}

/*
func (obj *ITaskbarList3) SetProgressState(hwnd HWND, state int) HRESULT {
	ret, _, _ := syscall.Syscall(obj.LpVtbl.SetProgressState, 3,
		uintptr(unsafe.Pointer(obj)),
		uintptr(hwnd),
		uintptr(state))
	return HRESULT(ret)
}
*/
func (g *goGenerator) generateInterfaceFunction(ii *InterfaceInfo, fi *FunctionInfo) {
	fn := fi.Function

	returnType := fi.ReturnType
	s := ""
	for _, arg := range fi.Args {
		s += fmt.Sprintf("%s %s, ", arg.Name, arg.TypeName)
	}
	s = strings.TrimSuffix(s, ", ")

	g.ws("\nfunc (i *%s) %s(%s) %s {\n", ii.Name(), fi.Name, s, returnType)

	// 	ret, _, _ := syscall.Syscall12(createWindowEx.Addr(), 12,
	// +1 because first argument is pointer to the COM object
	nArgs := len(fn.Params) + 1
	sysName, sysArgsCount := syscallFuncForArgCount(nArgs)
	if returnType != "" {
		g.ws("ret, _, _ := ")
	} else {
		g.ws("_, _, _ = ")
	}
	g.ws("%s(i.Vtbl.%s, %d,\n", sysName, fi.Name, nArgs)
	g.ws("uintptr(unsafe.Pointer(i)),\n")

	for _, arg := range fi.Args {
		if dePointerType(arg.TypeName) != "" {
			g.ws("uintptr(unsafe.Pointer(%s)),\n", arg.Name)
		} else {
			g.ws("uintptr(%s),\n", arg.Name)
		}
	}
	nLeftOver := sysArgsCount - nArgs
	for nLeftOver > 0 {
		g.ws("0,\n")
		nLeftOver--
	}
	g.ws(")\n")
	if returnType != "" {
		if returnType == "bool" {
			g.ws("return ret != 0\n")
		} else {
			g.ws("return %s(ret)\n", returnType)
		}
	}

	g.ws("\n}\n")
}

func (g *goGenerator) generateInterface(ii *InterfaceInfo) {
	if ii.WasGenerated {
		return
	}
	ii.WasGenerated = true

	i := ii.Definition.Interface

	// generate ID like
	// IID_ITaskbarList3 = IID{0xea1afb91, 0x9e28, 0x4b86, [8]byte{0x90, 0xe9, 0x9e, 0x9f, 0x8a, 0x5e, 0xef, 0xaf}}
	// from "{ea1afb91-9e28-4b86-90e9-9e9f8a5eefaf}

	a := splitGUID(i.ID)
	g.ws("var IID_%s = IID{%s, %s, %s, [8]byte{%s, %s, %s, %s, %s, %s, %s, %s}}\n\n", i.Name, a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10])

	// generate interface vtable
	vtableName := i.Name + "Vtbl"
	g.ws("type %s struct {\n", vtableName)
	if i.BaseInterface != "" {
		g.ws("\t%sVtbl\n", i.BaseInterface)
	}
	for _, m := range i.API {
		methodName := m.Name
		g.ws("%s uintptr\n", methodName)
	}
	g.ws("}\n\n")

	g.ws("type %s struct {\n", i.Name)
	if i.BaseInterface != "" {
		g.ws("\t%s\n", i.BaseInterface)
	}
	g.ws("\tVtbl *%s\n", vtableName)
	g.ws("}\n\n")

	for _, method := range ii.Methods {
		g.generateInterfaceFunction(ii, method)
	}
	// TODO: generate function to create the type
	//panic("NYI")
}

func dePointerType(typeName string) string {
	// TODO: should add a way to fully de-sugar types
	// this manually
	switch typeName {
	case "LPWSTR", "LPCWSTR":
		return typeName
	}
	if typeName[0] == '*' {
		return typeName[1:]
	}
	n := len(typeName)
	if typeName[n-1] == '*' {
		return typeName[:n-1]
	}
	return ""
}

func (g *goGenerator) generate() {
	var sourceFiles []string
	for sf := range g.sourceFiles {
		sourceFiles = append(sourceFiles, sf)
	}
	sort.Strings(sourceFiles)
	for _, name := range sourceFiles {
		sf := g.sourceFiles[name]
		g.generateSourceFile(sf)
		removeUnusedImportsInFile(sf.Path())
		gofmtFile(sf.Path())
		if printGeneratedFile {
			dumpFile(sf.Path())
		} else {
			fmt.Printf("Generated %s\n", sf.Path())
		}
	}
}

var whitelist = map[string]bool{
	"fast_utf8_to_utf16.go": true,
}

func shouldWhiteList(name string) bool {
	name = strings.ToLower(name)

	if whitelist[name] {
		return true
	}

	// all *util.go files are hand-written
	if strings.Contains(name, "util.go") {
		return true
	}

	// *_test.go files are also hand-written
	if strings.Contains(name, "_test.go") {
		return true
	}

	return false
}

func goDeleteExisting() {
	// delete all files in w except those that are hand-written (e.g. util.go)
	dir := "w"

	files, err := ioutil.ReadDir(dir)
	must(err)
	for _, fi := range files {
		if !fi.Mode().IsRegular() {
			continue
		}
		name := fi.Name()
		if shouldWhiteList(name) {
			continue
		}
		path := filepath.Join(dir, name)
		err = os.Remove(path)
		must(err)
	}
}

func tryCompile() {
	cmd := exec.Command("go", "build", "-o", "testw.exe", `github.com\kjk\winapigen\testw`)
	out, err := cmd.CombinedOutput()
	_ = os.Remove("testw.exe")
	if err != nil {
		fmt.Printf("Compilation failed with:\n%s\n", string(out))
	}
}

func genGo() {
	fmt.Printf("Starting gen go\n")
	out = os.Stdout
	goDeleteExisting()

	timeStart := time.Now()
	parsedFiles, err := parseApiMonitorData()
	must(err)
	fmt.Printf("Parsed %d files in %s\n", len(parsedFiles), time.Since(timeStart))

	timeStart = time.Now()
	buildIndex(parsedFiles)
	fmt.Printf("Built index in %s. %d functions, %d types, %d interfaces\n", time.Since(timeStart), len(allFunctions), len(allTypes), len(allInterfaces))

	g := newGoGenerator()
	functions := []string{"CoGetClassObject"}
	for _, f := range functions {
		g.addFunction(f)

	}
	//g.addFunction("CreateWindowExW")
	//g.addFunction("FileTimeToSystemTime")
	//g.addFunction("TzSpecificLocalTimeToSystemTime")
	//g.addFunction("GetSystemTimeAsFileTime")
	interfaces := []string{"IClassFactory", "IShellLinkW", "IPersistFile"}
	for _, i := range interfaces {
		g.addInterface(i)
	}
	g.generate()
	tryCompile()
}
