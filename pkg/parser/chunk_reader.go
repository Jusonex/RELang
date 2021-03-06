package parser

import (
	"strconv"

	"github.com/Jusonex/RELang/pkg/model"
	log "github.com/sirupsen/logrus"
)

// Type for creating a chunk model from the parsed AST
type ChunkReader struct {
	*BaseRELangListener

	ContextStack *GeneratorContextStack

	State struct {
		Chunk    *model.Chunk
		Class    *model.Class
		Function *model.Function
		Variable *model.Variable

		VariableType string
	}
}

// Constructs a new chunk reader
func NewChunkReader() *ChunkReader {
	s := new(ChunkReader)

	s.State.Chunk = model.NewChunk()
	s.ContextStack = NewGeneratorContextStack()

	return s
}

func (s *ChunkReader) EnterClassDeclaration(ctx *ClassDeclarationContext) {
	s.State.Class = model.NewClass(ctx.Name(0).GetText())

	for i, baseClass := range ctx.AllName() {
		if i > 0 {
			s.State.Class.AddBaseClass(baseClass.GetText())
		}
	}

	s.ContextStack.Push(CONTEXT_CLASS_DECL)
}

func (s *ChunkReader) ExitClassDeclaration(ctx *ClassDeclarationContext) {
	s.State.Chunk.AddClass(s.State.Class)

	s.ContextStack.Pop(CONTEXT_CLASS_DECL)
	s.State.Class = nil
}

func (s *ChunkReader) EnterFunctionDeclaration(ctx *FunctionDeclarationContext) {
	s.State.Function = model.NewFunction(ctx.Name().GetText())
	s.ContextStack.Push(CONTEXT_FUNCTION_DECL)
}

func (s *ChunkReader) ExitFunctionDeclaration(ctx *FunctionDeclarationContext) {
	s.ContextStack.Pop(CONTEXT_FUNCTION_DECL)

	inClass := s.ContextStack.Contains(CONTEXT_CLASS_DECL)
	if inClass {
		s.State.Class.AddFunction(s.State.Function)
	} else {
		if s.State.Function.MemoryAddress == nil {
			log.Fatal("global functions are required to have an address")
		}
		s.State.Chunk.AddFunction(s.State.Function)
	}

	s.State.Function = nil
}

func (s *ChunkReader) EnterFunctionModifier(ctx *FunctionModifierContext) {
	if s.ContextStack.Top() != CONTEXT_FUNCTION_DECL {
		// TODO: Log error
	}

	s.State.Function.Modifier = ctx.GetText()
}

func (s *ChunkReader) EnterFunctionReturnType(ctx *FunctionReturnTypeContext) {
	if s.ContextStack.Top() != CONTEXT_FUNCTION_DECL {
		// TODO: Log error
	}

	s.State.Function.ReturnType = ctx.GetText()
}

func (s *ChunkReader) EnterCallingConvention(ctx *CallingConventionContext) {
	if s.ContextStack.Top() != CONTEXT_FUNCTION_DECL {
		// TODO: Log error
	}

	s.State.Function.CallingConvention = ctx.GetText()
}

func (s *ChunkReader) ExitFunctionParameter(ctx *FunctionParameterContext) {
	s.State.Function.AddParameter(ctx.Name().GetText(), s.State.VariableType)
}

func (s *ChunkReader) EnterVariableDeclaration(ctx *VariableDeclarationContext) {
	s.State.Variable = model.NewVariable()

	s.ContextStack.Push(CONTEXT_VARIABLE_DECL)
}

func (s *ChunkReader) ExitVariableDeclaration(ctx *VariableDeclarationContext) {
	s.State.Variable.Name = ctx.Name().GetText()
	s.State.Variable.Type = s.State.VariableType

	inClass := s.ContextStack.Contains(CONTEXT_CLASS_DECL)
	if inClass {
		// Check if first offset is defined
		if len(s.State.Class.Variables) == 0 && s.State.Variable.MemoryOffset == nil {
			log.WithFields(log.Fields{"line": ctx.GetStart().GetLine()}).Fatal("the first variable offset needs to be explicitly defined")
		}

		s.State.Class.AddVariable(s.State.Variable)
	} else {
		if s.State.Variable.MemoryOffset == nil {
			log.WithFields(log.Fields{"line": ctx.GetStart().GetLine()}).Fatal("no memory address given for global variable")
		}

		s.State.Chunk.AddVariable(s.State.Variable)
	}

	s.ContextStack.Pop(CONTEXT_VARIABLE_DECL)
	s.State.Variable = nil
}

func (s *ChunkReader) EnterMemoryAddress(ctx *MemoryAddressContext) {
	// Memory address might not be supplied, so just ignore
	// TODO: Empty field can be removed from the grammer (=> ?-notation)
	if ctx.HexInteger() == nil {
		return
	}

	addressStr := ctx.HexInteger().GetText()
	memoryAddress, err := strconv.ParseUint(addressStr[2:], 16, 64)
	// TODO: Handle error
	if err != nil {
		panic(err)
	}

	if s.ContextStack.Top() == CONTEXT_FUNCTION_DECL {
		s.State.Function.MemoryAddress = &memoryAddress
	} else if s.ContextStack.Top() == CONTEXT_VARIABLE_DECL {
		s.State.Variable.MemoryOffset = &memoryAddress
	} else {
		log.Fatal("internal compiler error: unknown context in EnterMemoryAddress")
	}
}

func (s *ChunkReader) EnterPointer(ctx *PointerContext) {
	s.State.VariableType = ctx.Name().GetText() + "*"
}

func (s *ChunkReader) EnterPrimitiveType(ctx *PrimitiveTypeContext) {
	s.State.VariableType = ctx.GetText()
}

func (s *ChunkReader) EnterRawExpression(ctx *RawExpressionContext) {
	code := ctx.RawBlock().GetText()

	// Remove starting and ending ``` from the string
	code = code[3 : len(code)-3]

	// TODO: Fix indentation

	rawBlock := model.NewRawBlock(code)

	if s.ContextStack.Contains(CONTEXT_CLASS_DECL) {
		s.State.Class.AddRawBlock(rawBlock)
	} else {
		s.State.Chunk.AddRawBlock(rawBlock)
	}
}
