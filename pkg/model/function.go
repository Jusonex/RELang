package model

import "fmt"

type Parameter struct {
	Name string
	Type string
}

type FunctionModifier = string
type CallingConvention = string

type Function struct {
	Name              string
	ReturnType        string
	Modifier          FunctionModifier
	Params            []Parameter
	CallingConvention CallingConvention
	MemoryAddress     *uint64
}

func NewFunction(name string) *Function {
	return &Function{Name: name}
}

func NewFunctionPad(memoryOffset uint64) *Function {
	addr := memoryOffset

	return &Function{
		Name:              fmt.Sprintf("vpad_%x", memoryOffset),
		ReturnType:        "void",
		Modifier:          "virtual",
		CallingConvention: "",
		MemoryAddress:     &addr,
	}
}

func (s *Function) AddParameter(name string, parameterType string) {
	s.Params = append(s.Params, Parameter{Name: name, Type: parameterType})
}
