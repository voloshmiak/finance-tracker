package usecase

import "context"

type AssistantEngine interface {
	Ask(ctx context.Context, question, model string) (string, error)
}

type AskAssistant struct {
	engine AssistantEngine
}

func NewAskAssistant(engine AssistantEngine) *AskAssistant {
	return &AskAssistant{engine: engine}
}

func (u *AskAssistant) Execute(ctx context.Context, question, model string) (string, error) {
	return u.engine.Ask(ctx, question, model)
}
