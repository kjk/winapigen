package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
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

// information needed to generate info for a single module
// (aka dll).
type goSourceFile struct {
	// can be nil for interfaces
	module *Module

	// name of the .go file where we'll write generated info
	sourceFileName string

	functionsToGenerate []*FunctionInfo
	typesToGenerate     []*TypeInfo

	// interfaces don't come from DLLs but we reuse the
	// concept
	interfacesToGenerate []*InterfaceInfo
}

func (i *goSourceFile) Path() string {
	panicIf(i.sourceFileName == "")
	return filepath.Join("generated", i.sourceFileName)
}

// goGenerator keeps info needed for generating Go files
// we generate one .go file per module
type goGenerator struct {
	// maps module (dll) name to info for generating .go
	// file for that module
	modules map[string]*goSourceFile

	// we keep track of which const values have already been generated
	// because they the same constant might belong to multiple enums / sets
	generatedConsts map[string]struct{}

	// the file we're currently writing to
	w io.Writer
}

func newGoGenerator() *goGenerator {
	return &goGenerator{
		modules:         map[string]*goSourceFile{},
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

func (g *goGenerator) rememberType(vi *TypeInfo) {
	mi := g.getOrMakeModuleInfo(vi.Module)
	mi.typesToGenerate = append(mi.typesToGenerate, vi)
}

func (g *goGenerator) rememberInterface(ii *InterfaceInfo) {
	mi := g.getOrMakeModuleInfoForInterface(ii)
	mi.interfacesToGenerate = append(mi.interfacesToGenerate, ii)
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
		iname := g.addInterface(typeName)
		if iname != "" {
			return iname
		}
	}

	panicIf(vi == nil, "didn't find type with name '%s'", typeName)

	if vi.WasAdded {
		return vi.Name
	}
	vi.WasAdded = true

	v := vi.Variable
	if v.Type == typeVoid {
		vi.Name = "void"
		return vi.Name
	}

	if v.Type == typeModuleHandle {
		vi.Name = "HANDLE"
		return vi.Name
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
			f := &NameAndType{
				Name:     t.Name,
				TypeName: g.addType(t.Type, nil),
			}
			vi.Fields = append(vi.Fields, f)
		}
		g.rememberType(vi)
		vi.Name = v.Name
		return vi.Name
	}

	if v.Type == typePointer {
		// we don't add pointers
		vi.Name = "*" + g.addType(v.Base, nil)
		return vi.Name
	}

	if v.Type == typeArray {
		baseTypeName := g.addType(v.Base, nil)
		vi.Name = fmt.Sprintf("[%s]%s", v.Count, baseTypeName)
		return vi.Name
	}

	if v.Type == typeInteger {
		vi.Name = desugarInteger(v)
		return vi.Name
	}

	if v.Type == typeInterface {
		panic("interface NYI")
	}

	if v.Flag != nil {
		g.rememberType(vi)
		//fmt.Printf("Base for flag %s: %s\n", v.Name, v.Base)
		vi.Name = g.addType(v.Base, nil)
		return vi.Name
	}

	if v.Enum != nil {
		g.rememberType(vi)
		//fmt.Printf("Base for enum %s: %s\n", v.Name, v.Base)
		vi.Name = g.addType(v.Base, nil)
		return vi.Name
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
			vi.Name = v.Name
			g.rememberType(vi)
			return v.Name
		}
		return g.addType(v.Base, nil)
	}

	panic(fmt.Sprintf("unknown type %s", v.Type))
}

func (g *goGenerator) getOrMakeModuleInfoForInterface(ii *InterfaceInfo) *goSourceFile {
	name := ii.goSourceFileName()
	return g.getOrMakeModuleInfoForName(name)
}

func (g *goGenerator) getOrMakeModuleInfoForName(sourceFileName string) *goSourceFile {
	panicIf(sourceFileName == "", "name is empty")
	mi := g.modules[sourceFileName]
	if mi == nil {
		mi = &goSourceFile{
			sourceFileName: sourceFileName,
		}
		g.modules[sourceFileName] = mi
	}
	return mi
}

func (g *goGenerator) getOrMakeModuleInfo(mod *Module) *goSourceFile {
	if mod == nil {
		return g.getOrMakeModuleInfoForName(noModule)
	}
	return g.getOrMakeModuleInfoForName(mod.goSourceFileName())
}

// types with name "IBindCtx*" is a pointer to an interface IBindCtx
// returns interface info or nil if s not a pointer to an interface
func findInterfaceByNameOrPointer(s string) *InterfaceInfo {
	for strings.HasSuffix(s, "*") {
		s = strings.TrimSuffix(s, "*")
	}
	return allInterfaces[s]
}

func (g *goGenerator) addInterface(name string) string {
	ii := findInterfaceByNameOrPointer(name)
	if ii == nil {
		return ""
	}
	if ii.WasAdded {
		return ii.Name()
	}
	ii.WasAdded = true

	mod := g.getOrMakeModuleInfoForInterface(ii)
	mod.interfacesToGenerate = append(mod.interfacesToGenerate, ii)
	i := ii.Definition.Interface

	for _, method := range i.API {
		api := method
		name := api.Name
		fi := &FunctionInfo{
			SourceFile: ii.Declaration.SourceFile,
			Function:   api,
			Name:       name,
			Module:     mod.module,
		}

		g.buildFunctionInfo(fi)
		ii.Methods = append(ii.Methods, fi)
	}

	if i.BaseInterface != "" {
		baseName := g.addInterface(i.BaseInterface)
		panicIf(baseName == "", "didn't add interface '%s'", i.BaseInterface)
	}

	for _, v := range i.Variable {
		g.addType(v.Name, nil)
	}
	return ii.Name()
}

func (g *goGenerator) buildFunctionInfo(fi *FunctionInfo) {
	ret := fi.Function.Return
	fi.ReturnType = desugarReturnType(g.addType(ret.Type, nil))

	for _, arg := range fi.Function.Params {
		farg := &NameAndType{
			Name:     arg.Name,
			TypeName: g.addType(arg.Type, nil),
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

	mi := g.getOrMakeModuleInfo(fi.Module)
	mi.functionsToGenerate = append(mi.functionsToGenerate, fi)
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
		g.ws("type %s %s\n\n", v.Name, v.Type)
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

func (g *goGenerator) generateModuleFunctionBodies(mi *goSourceFile) {
	if len(mi.functionsToGenerate) == 0 {
		return
	}
	for _, fi := range mi.functionsToGenerate {
		g.generateFunction(fi)
	}
}

func (g *goGenerator) generateModuleFunctionVariables(mi *goSourceFile) {
	if len(mi.functionsToGenerate) == 0 {
		return
	}
	mod := mi.module
	panicIf(mod == nil, "mi.module should not be nil")
	modName := mod.moduleName()
	dllName := mod.dllName()

	g.ws("var (\n")

	// global variable that is a handle to dll
	g.ws("lib%s *windows.LazyDLL\n\n", modName)

	// for each function from dll that we use, a global variable
	// for the function address, like this:
	// abortDoc               *windows.LazyProc
	for _, fi := range mi.functionsToGenerate {
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

	for _, fi := range mi.functionsToGenerate {
		g.ws("\t%s = lib%s.NewProc(\"%s\")\n", fi.GoVarName(), modName, fi.Name)
	}

	g.ws("}\n")
}

func (g *goGenerator) generateModuleInterfaces(mi *goSourceFile) {
	if len(mi.interfacesToGenerate) == 0 {
		return
	}
	panic("interfaces NYI")
}

func (g *goGenerator) generateModule(mi *goSourceFile) {
	fileName := mi.sourceFileName
	fmt.Printf("Generating module %s\n", fileName)
	var err error
	g.w, err = os.Create(mi.Path())
	must(err)
	defer func() {
		f := g.w.(*os.File)
		err := f.Close()
		must(err)
		g.w = nil
	}()

	g.ws(fileHdr)

	g.generateModuleFunctionVariables(mi)

	for _, ti := range mi.typesToGenerate {
		g.generateType(ti)
	}

	g.generateModuleFunctionBodies(mi)

	g.generateModuleInterfaces(mi)

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
	case "LPVOID":
		return "unsafe.Pointer"
	case "LPCTSTR":
		return "*uint16"
	case "int8", "int16", "int32", "int64":
		return tp
	case "uint8", "uint16", "uint32", "uint64":
		return tp
	case "WCHAR":
		return tp
	case "UINT_PTR":
		return "uintptr"
	case "int":
		// yes, it's strange but that's what it seems to be
		return "int32"
	case "ModuleHandle":
		return "HANDLE"
	case "Guid", "GUID":
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
		if isPointerType(arg.TypeName) {
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

func (g *goGenerator) generateInterface(name string) {
	ii := findInterface(name)
	if ii.WasGenerated {
		return
	}
	ii.WasGenerated = true

	i := ii.Definition.Interface
	baseName := i.BaseInterface
	if baseName != "" {
		g.generateInterface(baseName)
	}

	// TODO: generate vtable
	// TODO: generate functions wrapping vtable
	// TODO: generate function to create the type
	panic("NYI")
}

func isPointerType(typeName string) bool {
	return typeName[0] == '*'
}

func (g *goGenerator) generate() {
	var modules []string
	for mod := range g.modules {
		modules = append(modules, mod)
	}
	sort.Strings(modules)
	for _, name := range modules {
		mi := g.modules[name]
		g.generateModule(mi)
		gofmtFile(mi.Path())
		dumpFile(mi.Path())
	}
}

func genGo() {
	fmt.Printf("Starting gen go\n")
	out = os.Stdout

	timeStart := time.Now()
	parsedFiles, err := parseApiMonitorData()
	must(err)
	fmt.Printf("Parsed %d files in %s\n", len(parsedFiles), time.Since(timeStart))

	timeStart = time.Now()
	buildIndex(parsedFiles)
	fmt.Printf("Built index in %s. %d functions, %d variables, %d interfaces\n", time.Since(timeStart), len(allFunctions), len(allTypes), len(allInterfaces))

	g := newGoGenerator()
	g.addFunction("CreateWindowExW")
	//g.addFunction("FileTimeToSystemTime")
	//g.addFunction("TzSpecificLocalTimeToSystemTime")
	//g.addFunction("GetSystemTimeAsFileTime")
	//g.addInterface("IBindHost")
	g.generate()
}
