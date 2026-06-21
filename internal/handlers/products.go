package handlers

import (
	"database/sql"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterProducts(se *core.ServeEvent, app *pocketbase.PocketBase) {
	se.Router.GET("/api/products", func(re *core.RequestEvent) error {
		records, err := app.FindRecordsByFilter(
			"products",
			"visible = true",
			"-created",
			50,
			0,
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

	se.Router.GET("/api/products/{slug}", func(re *core.RequestEvent) error {
		slug := re.Request.PathValue("slug")
		record, err := app.FindFirstRecordByFilter(
			"products",
			"slug = {:slug}",
			map[string]any{"slug": slug},
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return re.String(404, `{"error":"product not found"}`)
			}
			return re.String(500, `{"error":"`+err.Error()+`"}`)
		}
		return re.JSON(200, record.PublicExport())
	})
}
