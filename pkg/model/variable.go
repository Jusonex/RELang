package model

import "fmt"

type Variable struct {
	Name         string
	Type         string
	MemoryOffset *uint64
}

func NewVariable() *Variable {
	return &Variable{}
}

func NewVariablePad(memoryOffset uint64, size uint64) *Variable {
	addr := memoryOffset

	return &Variable{
		// HACK: Size was embedded in name, fix this one day!
		Name:         fmt.Sprintf("pad_%x[%d]", memoryOffset, size),
		Type:         "char",
		MemoryOffset: &addr,
	}
}
