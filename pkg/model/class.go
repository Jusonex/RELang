package model

// Represents a class within a chunk
type Class struct {
	Name             string
	BaseClasses      []string
	Variables        []*Variable
	Functions        []*Function
	VirtualFunctions []*Function
	StaticFunctions  []*Function
	RawBlocks        []*RawBlock
}

// Constructs a new class by the given name
func NewClass(name string) *Class {
	return &Class{Name: name}
}

// Adds a base class to the chunk
func (s *Class) AddBaseClass(name string) {
	s.BaseClasses = append(s.BaseClasses, name)
}

// Adds a function (method) to the class
func (s *Class) AddFunction(function *Function) {
	switch function.Modifier {
	case "virtual":
		s.VirtualFunctions = append(s.VirtualFunctions, function)
		break
	case "static":
		s.StaticFunctions = append(s.StaticFunctions, function)
		break
	default:
		s.Functions = append(s.Functions, function)
	}
}

// Adds a variable (attribute) to the class
func (s *Class) AddVariable(variable *Variable) {
	s.Variables = append(s.Variables, variable)
}

// Adds a raw block to the class
func (s *Class) AddRawBlock(rawBlock *RawBlock) {
	s.RawBlocks = append(s.RawBlocks, rawBlock)
}

// Calculates the size of a class. This method has to be called after MemoryOffets are filled
// If virtual-offsetting is applied through transformers, vtable ptr is implicitly included
func (s *Class) GetSize() uint64 {
	numVariables := len(s.Variables)
	if numVariables == 0 {
		return uint64(0)
	}

	lastVariable := s.Variables[numVariables-1]
	return *lastVariable.MemoryOffset + lastVariable.Size
}

// HasVirtualMembers returns true if the class has virtual methods, false otherwise
func (s *Class) HasVirtualMembers() bool {
	return len(s.VirtualFunctions) > 0
}
