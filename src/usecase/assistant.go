package usecase

import "context"

// AssistantEngine answers a natural-language finance question.
type AssistantEngine interface {
	Ask(ctx context.Context, question string) (string, error)
}

type Assistant struct {
	engine AssistantEngine
}

func NewAssistant(engine AssistantEngine) *Assistant {
	return &Assistant{engine: engine}
}

func (u *Assistant) Ask(ctx context.Context, question string) (string, error) {
	return u.engine.Ask(ctx, question)
}
