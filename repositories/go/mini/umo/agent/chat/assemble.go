package chat

type Assembler interface {
	AssembleChatPrompt(args ...any) *AssembledPrompt
}

type AssembledPrompt struct{}
