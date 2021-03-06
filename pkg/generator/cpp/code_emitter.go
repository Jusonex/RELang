package cpp

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Jusonex/RELang/pkg/model"
)

type CodeEmitter struct {
	OutputFile *os.File
	Writer     *bufio.Writer

	TabIndex      int
	PublicContext bool
}

func NewCodeEmitter(path string) *CodeEmitter {
	e := new(CodeEmitter)
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

func (s *CodeEmitter) Close() {
	s.Writer.Flush()
	s.OutputFile.Close()
}

func (s *CodeEmitter) EmitHeader() {
	s.Writer.WriteString("// DO NOT EDIT. THIS FILE WAS GENERATED BY THE RELANG COMPILER\n")
}

func (s *CodeEmitter) EmitLineComment(comment string) {
	s.EmitIndentation()
	s.Writer.WriteString("// " + comment + "\n")
}

func (s *CodeEmitter) EmitSeparator() {
	s.Writer.WriteString(";")
}

func (s *CodeEmitter) EmitIndentation() {
	s.Writer.WriteString(strings.Repeat("    ", s.TabIndex))
}

func (s *CodeEmitter) EmitLine(line string, separator bool) {
	s.EmitIndentation()

	if separator {
		s.Writer.WriteString(line + ";\n")
	} else {
		s.Writer.WriteString(line + "\n")
	}
}

func (s *CodeEmitter) EmitPublicBlockIfNecessary() {
	if s.PublicContext {
		return
	}

	s.TabIndex = s.TabIndex - 1 // temporally step back tab index
	s.EmitLine("public:", false)
	s.TabIndex = s.TabIndex + 1
	s.PublicContext = true
}

func (s *CodeEmitter) EmitPrivateBlockIfNecessary() {
	if !s.PublicContext {
		return
	}

	s.TabIndex = s.TabIndex - 1 // temporally step back tab index
	s.EmitLine("private:", false)
	s.TabIndex = s.TabIndex + 1
	s.PublicContext = false
}

func (s *CodeEmitter) EmitAccessBlock(public bool) {
	if public {
		s.EmitPublicBlockIfNecessary()
	} else {
		s.EmitPrivateBlockIfNecessary()
	}
}

func (s *CodeEmitter) EmitIncludeGuard() {
	s.EmitLine("#pragma once", false)
}

func (s *CodeEmitter) EmitIncludeStatement(includePath string, relativeInclude bool) {
	if relativeInclude {
		s.EmitLine("#include \""+includePath+"\"", false)
	} else {
		s.EmitLine("#include <"+includePath+">", false)
	}
}

func (s *CodeEmitter) EmitForwardDeclaration(className string) {
	s.EmitLine("class "+className, true)
}

func (s *CodeEmitter) EmitClassDeclarationStart(className string, baseClasses []string) {
	s.EmitLine("#pragma pack(push)", false)
	s.EmitLine("#pragma pack(1)", false)

	if len(baseClasses) > 0 {
		baseClassEnumeration := strings.Join(baseClasses, ", public ")
		s.EmitLine(fmt.Sprintf("class %s : public %s\n{", className, baseClassEnumeration), false)
	} else {
		s.EmitLine(fmt.Sprintf("class %s\n{", className), false)
	}
	s.TabIndex = s.TabIndex + 1

	s.EmitPublicBlockIfNecessary()
}

func (s *CodeEmitter) EmitClassDeclarationEnd() {
	s.TabIndex = s.TabIndex - 1
	s.EmitLine("}", true)
	s.EmitLine("#pragma pack(pop)", false)
}

func (s *CodeEmitter) EmitClassSizeAssertion(className string, expectedSize uint64) {
	s.EmitLine(fmt.Sprintf("static_assert(sizeof(%s) == 0x%X, \"Unexpected class size\")", className, expectedSize), true)
}

func (s *CodeEmitter) EmitFunctionDeclaration(function *model.Function, inClass bool) {
	if inClass {
		s.EmitAccessBlock(function.Public)
	}

	s.EmitLine(fmt.Sprintf("inline %s %s(%s)", function.ReturnType, function.Name, FunctionParametersToString(function.Params)), false)
	s.EmitLine("{", false)
	s.TabIndex = s.TabIndex + 1

	params := function.Params
	if inClass {
		params = append([]model.Parameter{model.Parameter{Name: "this", Type: "decltype(this)"}}, params...)
	}

	s.EmitLine(fmt.Sprintf("using Func_t = %s(%s *)(%s)", function.ReturnType, function.CallingConvention, FunctionParameterTypesToString(params)), true)
	s.EmitLine(fmt.Sprintf("auto f = reinterpret_cast<Func_t>(0x%X)", *function.MemoryAddress), true)
	s.EmitLine(fmt.Sprintf("return f(%s)", FunctionParameterNamesToString(params, true)), true)

	s.TabIndex = s.TabIndex - 1
	s.EmitLine("}\n", false)
}

func (s *CodeEmitter) EmitVirtualFunctionDeclaration(function *model.Function) {
	s.EmitAccessBlock(function.Public)

	s.EmitLine(fmt.Sprintf("virtual %s %s(%s) = 0", function.ReturnType, function.Name, FunctionParametersToString(function.Params)), true)
}

func (s *CodeEmitter) EmitClassVariableDeclaration(variable *model.Variable) {
	s.EmitAccessBlock(variable.Public)

	s.EmitLine(fmt.Sprintf("%s %s; // offset 0x%X", variable.Type, variable.Name, *variable.MemoryOffset), false)
}

func (s *CodeEmitter) EmitGlobalVariableDeclaration(variable *model.Variable) {
	s.EmitLine(fmt.Sprintf("inline %s& %s = *(%s*)0x%X", variable.Type, variable.Name, variable.Type, *variable.MemoryOffset), true)
}
