package assistant

import (
	"context"
	"fmt"
	"maps"
	"time"

	"google.golang.org/genai"
)

const maxToolTurns = 8

var allowedModels = map[string]bool{
	"gemini-2.5-flash": true,
	"gemini-3.5-flash": true,
}

type Tool interface {
	Declaration() *genai.FunctionDeclaration
	Execute(ctx context.Context, args map[string]any) (map[string]any, error)
}

func systemPrompt() string {
	return `Ти — фінансовий асистент у застосунку обліку особистих витрат.
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

Таблиця "mcc_categories" — довідник категорій торговців:
  - mcc (int): код MCC
  - category (text): укрупнена категорія українською ('Продукти', 'Кафе і ресторани', 'Транспорт', ...)
  - title (text): детальний опис
  - is_spending (bool): чи це реальна витрата. FALSE для переказів, зняття готівки,
    фінансових операцій — це рух коштів, а НЕ споживання.
Щоб групувати витрати за категоріями, JOIN-ай transactions.mcc = mcc_categories.mcc.
Транзакції без MCC або яких немає в довіднику відноси до 'Інше'.

ВАЖЛИВО: коли рахуєш ВИТРАТИ (споживання), ВИКЛЮЧАЙ перекази й готівку:
  AND NOT EXISTS (SELECT 1 FROM mcc_categories mc WHERE mc.mcc = transactions.mcc AND mc.is_spending = false)
Перекази/готівку аналізуй окремо, лише якщо користувач прямо про них питає.

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
- Якщо даних справді немає — скажи про це чесно й запропонуй, що можна завантажити.

Сьогоднішня дата: ` + time.Now().Format("2006-01-02") + `

Мій дохід складає 1300$ на місяць.

`
}

type Assistant struct {
	client *genai.Client
	tools  map[string]Tool
	decls  []*genai.Tool
}

func New(apiKey string) (*Assistant, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Assistant{
		client: client,
		tools:  make(map[string]Tool),
	}, nil
}

func (a *Assistant) Register(tools ...Tool) {
	for _, tool := range tools {
		a.tools[tool.Declaration().Name] = tool
		a.decls = append(a.decls, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{tool.Declaration()},
		})
	}
}

func (a *Assistant) Ask(ctx context.Context, question, model string) (string, error) {
	if !allowedModels[model] {
		return "", fmt.Errorf("model %q not allowed; must be one of: %v",
			model, maps.Keys(allowedModels))
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt(), genai.RoleUser),
		Tools:             a.decls,
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

// dispatch routes a function call to its registered tool. Unknown tools and
// execution errors are reported back to the model so it can recover.
func (a *Assistant) dispatch(ctx context.Context, call *genai.FunctionCall) map[string]any {
	tool, ok := a.tools[call.Name]
	if !ok {
		return map[string]any{"error": fmt.Sprintf("unknown tool %q", call.Name)}
	}

	result, err := tool.Execute(ctx, call.Args)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return result
}
