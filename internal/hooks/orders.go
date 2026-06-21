package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterOrderHooks(app *pocketbase.PocketBase) {
	app.OnRecordCreate("orders").BindFunc(func(e *core.RecordEvent) error {
		return calculateOrderTotal(e)
	})
}

func calculateOrderTotal(e *core.RecordEvent) error {
	items, ok := e.Record.Get("items").([]any)
	if !ok {
		return e.Next()
	}

	var total float64
	for _, item := range items {
		if i, ok := item.(map[string]any); ok {
			price, _ := i["price"].(float64)
			qty, _ := i["quantity"].(float64)
			if qty == 0 {
				qty = 1
			}
			total += price * qty
		}
	}
	e.Record.Set("total", total)
	return e.Next()
}
