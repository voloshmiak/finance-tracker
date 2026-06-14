package usecase

import (
	"context"

	"finance-tracker/src/assistant"
)

// AssistantEngine answers a natural-language finance question and lists models.
type AssistantEngine interface {
	Ask(ctx context.Context, question, model string) (string, error)
	ListModels(ctx context.Context) ([]assistant.Model, error)
	DefaultModel() string
}

type Assistant struct {
	engine AssistantEngine
}

func NewAssistant(engine AssistantEngine) *Assistant {
	return &Assistant{engine: engine}
}

func (u *Assistant) Ask(ctx context.Context, question, model string) (string, error) {
	return u.engine.Ask(ctx, question, model)
}

func (u *Assistant) Models(ctx context.Context) ([]assistant.Model, error) {
	return u.engine.ListModels(ctx)
}

func (u *Assistant) DefaultModel() string {
	return u.engine.DefaultModel()
}
