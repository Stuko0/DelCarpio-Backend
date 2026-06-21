package hooks

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterAuthHooks(app *pocketbase.PocketBase) {
	app.OnRecordAfterCreateSuccess("users").BindFunc(func(e *core.RecordEvent) error {
		collection, err := app.FindCollectionByNameOrId("profiles")
		if err != nil {
			return e.Next()
		}

		profile := core.NewRecord(collection)
		profile.Set("user", e.Record.Id)
		profile.Set("name", e.Record.GetString("name"))
		profile.Set("email", e.Record.GetString("email"))

		if err := app.Save(profile); err != nil {
			return err
		}
		return e.Next()
	})
}
