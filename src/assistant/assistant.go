package assistant

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// maxToolTurns bounds the agentic loop so a misbehaving model cannot loop forever.
const maxToolTurns = 8

// Model is a Gemini model the user can pick from in the UI.
type Model struct {
	ID    string `json:"id"`    // e.g. "gemini-2.5-flash"
	Label string `json:"label"` // human-friendly display name
}

// ListModels queries the Gemini API for models that support generateContent.
func (a *Assistant) ListModels(ctx context.Context) ([]Model, error) {
	var models []Model
	for m, err := range a.client.Models.All(ctx) {
		if err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
		if !supportsGenerate(m) {
			continue
		}
		id := strings.TrimPrefix(m.Name, "models/")
		if !suitableForTools(id) {
			continue
		}
		label := m.DisplayName
		if label == "" {
			label = id
		}
		models = append(models, Model{ID: id, Label: label})
	}
	return models, nil
}

func supportsGenerate(m *genai.Model) bool {
	for _, a := range m.SupportedActions {
		if a == "generateContent" {
			return true
		}
	}
	return false
}

// suitableForTools keeps only stable Gemini chat models that support function
// calling. Embedding/TTS/image/vision-only and preview/experimental variants
// either lack tool support or use an incompatible request shape, which is what
// caused "gemini call" errors on non-flash models.
func suitableForTools(id string) bool {
	if !strings.HasPrefix(id, "gemini-") {
		return false
	}
	for _, bad := range []string{
		"embedding", "tts", "image", "vision", "live",
		"preview", "exp", "thinking", "learnlm",
	} {
		if strings.Contains(id, bad) {
			return false
		}
	}
	return true
}

// DefaultModel returns the model used when a request does not specify one.
func (a *Assistant) DefaultModel() string {
	return a.model
}

const systemPrompt = `Ти — фінансовий асистент у застосунку обліку особистих витрат.
Користувач ставить питання українською про свої витрати з банківських карток.

Дані лежать у таблиці PostgreSQL "transactions":
  - id (uuid)
  - bank (text): 'monobank' | 'credit_agricole' | 'abank'
  - operated_at (timestamptz): дата й час операції
  - description (text): опис операції
  - amount (bigint): сума у КОПІЙКАХ у валюті картки; ВІД'ЄМНА = витрата, ДОДАТНА = надходження
  - currency (text): валюта рахунку, напр. 'UAH'
  - mcc (int, може бути NULL): код категорії торговця
  - cashback, commission, balance_after (bigint, копійки, можуть бути NULL)
  - card_mask (text, NULL), last4 (char(4)): остання 4 цифри картки
  - created_at (timestamptz)

Таблиця "cards": last4, bank, label, currency, is_active.

Твоя робота:
- Відповідати на питання про витрати, аналізувати їх і ДАВАТИ ПОРАДИ щодо економії та бюджету.
- Коли користувач просить поради — НЕ відмовляйся. Спершу самостійно дослідь дані через run_sql
  (наприклад: найбільші категорії за MCC, динаміка по місяцях, найбільші окремі витрати,
  співвідношення витрат до надходжень), а потім дай конкретні, персоналізовані рекомендації
  на основі побачених цифр. Не питай дозволу заглянути в дані — просто роби це.

Як працювати з даними:
- ВИКОРИСТОВУЙ інструмент run_sql скільки потрібно — пиши лише SELECT-запити.
- Суми зберігаються в копійках: діли на 100, щоб показати гривні. Витрати — додатними числами за модулем.
- Для категорій орієнтуйся на MCC та текст у description.

Стиль відповіді:
- Українською, дружньо і змістовно. Не обмежуйся одним рядком — давай контекст, порівняння,
  тенденції, і де доречно — список конкретних дій.
- Підкріплюй висновки реальними цифрами з бази.
- Якщо даних справді немає — скажи про це чесно й запропонуй, що можна завантажити.`

// Assistant answers natural-language questions about the user's finances by
// letting Gemini call a read-only SQL tool.
type Assistant struct {
	client  *genai.Client
	model   string
	sqlTool *SQLTool
}

func New(ctx context.Context, apiKey, model string, sqlTool *SQLTool) (*Assistant, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Assistant{client: client, model: model, sqlTool: sqlTool}, nil
}

func sqlToolDeclaration() *genai.Tool {
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{{
			Name:        "run_sql",
			Description: "Виконує SELECT-запит до бази даних транзакцій і повертає рядки у форматі JSON.",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"query": {
						Type:        genai.TypeString,
						Description: "Один SELECT-запит PostgreSQL. Без крапки з комою в кінці, без мутацій.",
					},
				},
				Required: []string{"query"},
			},
		}},
	}
}

// Ask runs the agentic loop: the model may call run_sql several times before
// producing a final natural-language answer. An empty model uses the default.
func (a *Assistant) Ask(ctx context.Context, question, model string) (string, error) {
	if model == "" {
		model = a.model
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		Tools:             []*genai.Tool{sqlToolDeclaration()},
	}

	contents := []*genai.Content{
		genai.NewContentFromText(question, genai.RoleUser),
	}

	for turn := 0; turn < maxToolTurns; turn++ {
		resp, err := a.client.Models.GenerateContent(ctx, model, contents, config)
		if err != nil {
			return "", fmt.Errorf("gemini call: %w", err)
		}

		calls := resp.FunctionCalls()
		if len(calls) == 0 {
			return resp.Text(), nil // final answer
		}

		// Append the model's tool-call turn, then answer each call.
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			contents = append(contents, resp.Candidates[0].Content)
		}

		responseParts := make([]*genai.Part, 0, len(calls))
		for _, call := range calls {
			result := a.dispatch(ctx, call)
			responseParts = append(responseParts, genai.NewPartFromFunctionResponse(call.Name, result))
		}
		contents = append(contents, genai.NewContentFromParts(responseParts, genai.RoleUser))
	}

	return "", fmt.Errorf("assistant exceeded %d tool turns without answering", maxToolTurns)
}

// dispatch executes a single tool call and returns the response payload.
func (a *Assistant) dispatch(ctx context.Context, call *genai.FunctionCall) map[string]any {
	if call.Name != "run_sql" {
		return map[string]any{"error": fmt.Sprintf("unknown tool %q", call.Name)}
	}

	query, _ := call.Args["query"].(string)
	rows, err := a.sqlTool.Run(ctx, query)
	if err != nil {
		// Feed the error back so the model can correct its query.
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"rows": rows}
}
