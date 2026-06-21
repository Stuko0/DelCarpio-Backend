package hooks

import (
	"regexp"
	"strings"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterProductHooks(app *pocketbase.PocketBase) {
	app.OnRecordCreate("products").BindFunc(func(e *core.RecordEvent) error {
		return validateProduct(e)
	})
	app.OnRecordUpdate("products").BindFunc(func(e *core.RecordEvent) error {
		return validateProduct(e)
	})
}

func validateProduct(e *core.RecordEvent) error {
	if e.Record.GetString("slug") == "" {
		e.Record.Set("slug", createSlug(e.Record.GetString("name")))
	}
	return e.Next()
}

func createSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "ñ", "n")
	slug = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
