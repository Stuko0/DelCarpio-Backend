package handlers

import (
	"database/sql"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterOrders(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.POST("/api/orders", func(re *core.RequestEvent) error {
		if re.Auth == nil {
			return re.String(401, `{"error":"unauthorized"}`)
		}

		data := struct {
			Items []map[string]any `json:"items"`
		}{}
		if err := re.BindBody(&data); err != nil {
			return re.String(400, `{"error":"invalid body"}`)
		}

		collection, err := app.FindCollectionByNameOrId("orders")
		if err != nil {
			return re.String(500, `{"error":"collection not found"}`)
		}

		record := core.NewRecord(collection)
		record.Set("user", re.Auth.Id)
		record.Set("items", data.Items)
		record.Set("status", "pending")

		if err := app.Save(record); err != nil {
			return re.String(500, `{"error":"`+err.Error()+`"}`)
		}
		return re.JSON(201, record.PublicExport())
	})

	se.Router.GET("/api/orders", func(re *core.RequestEvent) error {
		if re.Auth == nil {
			return re.String(401, `{"error":"unauthorized"}`)
		}

		records, err := app.FindRecordsByFilter(
			"orders",
			"user = {:user}",
			"-created",
			50,
			0,
			map[string]any{"user": re.Auth.Id},
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return re.JSON(200, []map[string]any{})
			}
			return re.String(500, `{"error":"`+err.Error()+`"}`)
		}
		result := make([]map[string]any, 0, len(records))
		for _, r := range records {
			result = append(result, r.PublicExport())
		}
		return re.JSON(200, result)
	})
}
