package generator

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Jusonex/RELang/pkg/model"
)

type CppCodeEmitter struct {
	OutputFile *os.File
	Writer     *bufio.Writer

	TabIndex      int
	PublicContext bool
}

func NewCppCodeEmitter(path string) *CppCodeEmitter {
	e := new(CppCodeEmitter)
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	e.OutputFile = file
	e.Writer = bufio.NewWriter(e.OutputFile)
	e.TabIndex = 0
	e.PublicContext = false

	return e
}

func (s *CppCodeEmitter) Close() {
	s.Writer.Flush()
	s.OutputFile.Close()
}

func (s *CppCodeEmitter) EmitHeader() {
	s.Writer.WriteString("// DO NOT EDIT. THIS FILE WAS GENERATED BY THE RELANG COMPILER\n\n")
}

func (s *CppCodeEmitter) EmitLineComment(comment string) {
	s.EmitIndentation()
	s.Writer.WriteString("// " + comment + "\n")
}

func (s *CppCodeEmitter) EmitSeparator() {
	s.Writer.WriteString(";")
}

func (s *CppCodeEmitter) EmitIndentation() {
	s.Writer.WriteString(strings.Repeat("    ", s.TabIndex))
}

func (s *CppCodeEmitter) EmitLine(line string, separator bool) {
	s.EmitIndentation()

	if separator {
		s.Writer.WriteString(line + ";\n")
	} else {
		s.Writer.WriteString(line + "\n")
	}
}

func (s *CppCodeEmitter) EmitPublicBlockIfNecessary() {
	if s.PublicContext {
		return
	}

	s.TabIndex = s.TabIndex - 1 // temporally step back tab index
	s.EmitLine("public:", false)
	s.TabIndex = s.TabIndex + 1
	s.PublicContext = true
}

func (s *CppCodeEmitter) EmitPrivateBlockIfNecessary() {
	if !s.PublicContext {
		return
	}

	s.TabIndex = s.TabIndex - 1 // temporally step back tab index
	s.EmitLine("private:", false)
	s.TabIndex = s.TabIndex + 1
	s.PublicContext = false
}

func (s *CppCodeEmitter) EmitAccessBlock(public bool) {
	if public {
		s.EmitPublicBlockIfNecessary()
	} else {
		s.EmitPrivateBlockIfNecessary()
	}
}

func (s *CppCodeEmitter) EmitIncludeStatement(includePath string) {
	s.EmitLine("#include \""+includePath+"\"", false)
}

func (s *CppCodeEmitter) EmitClassDeclarationStart(className string, baseClasses []string) {
	if len(baseClasses) > 0 {
		baseClassEnumeration := strings.Join(baseClasses, ", public ")
		s.EmitLine(fmt.Sprintf("class %s : public %s\n{", className, baseClassEnumeration), false)
	} else {
		s.EmitLine(fmt.Sprintf("class %s\n{", className), false)
	}
	s.TabIndex = s.TabIndex + 1

	s.EmitPublicBlockIfNecessary()
}

func (s *CppCodeEmitter) EmitClassDeclarationEnd() {
	s.TabIndex = s.TabIndex - 1
	s.EmitLine("}", true)
}

func (s *CppCodeEmitter) EmitFunctionDeclaration(function *model.Function, inClass bool) {
	if inClass {
		s.EmitAccessBlock(function.Public)
	}

	s.EmitLine(fmt.Sprintf("inline %s %s(%s)", function.ReturnType, function.Name, CppFunctionParametersToString(function.Params)), false)
	s.EmitLine("{", false)
	s.TabIndex = s.TabIndex + 1

	params := function.Params
	if inClass {
		params = append([]model.Parameter{model.Parameter{Name: "this", Type: "decltype(this)"}}, params...)
	}

	s.EmitLine(fmt.Sprintf("using Func_t = %s(%s *)(%s)", function.ReturnType, function.CallingConvention, CppFunctionParameterTypesToString(params)), true)
	s.EmitLine(fmt.Sprintf("auto f = reinterpret_cast<Func_t>(0x%x)", *function.MemoryAddress), true)
	s.EmitLine(fmt.Sprintf("return f(%s)", CppFunctionParameterNamesToString(params, true)), true)

	s.TabIndex = s.TabIndex - 1
	s.EmitLine("}\n", false)
}

func (s *CppCodeEmitter) EmitVirtualFunctionDeclaration(function *model.Function) {
	s.EmitAccessBlock(function.Public)

	s.EmitLine(fmt.Sprintf("virtual %s %s(%s) = 0", function.ReturnType, function.Name, CppFunctionParametersToString(function.Params)), true)
}

func (s *CppCodeEmitter) EmitVariableDeclaration(variable *model.Variable) {
	s.EmitAccessBlock(variable.Public)

	s.EmitLine(fmt.Sprintf("%s %s; // offset 0x%X", variable.Type, variable.Name, *variable.MemoryOffset), false)
}
