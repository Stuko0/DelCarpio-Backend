package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_4092854851")
		if err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(2, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text2560465762",
			"max": 0,
			"min": 0,
			"name": "slug",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": true,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(3, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text1843675174",
			"max": 0,
			"min": 0,
			"name": "description",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": false,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(4, []byte(`{
			"hidden": false,
			"id": "number3402113753",
			"max": null,
			"min": 10,
			"name": "price",
			"onlyInt": false,
			"presentable": false,
			"required": true,
			"system": false,
			"type": "number"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(5, []byte(`{
			"autogeneratePattern": "",
			"hidden": false,
			"id": "text105650625",
			"max": 0,
			"min": 0,
			"name": "category",
			"pattern": "",
			"presentable": false,
			"primaryKey": false,
			"required": true,
			"system": false,
			"type": "text"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(6, []byte(`{
			"hidden": false,
			"id": "file3309110367",
			"maxSelect": 1,
			"maxSize": 0,
			"mimeTypes": [
				"image/jpeg",
				"image/png"
			],
			"name": "image",
			"presentable": false,
			"protected": false,
			"required": true,
			"system": false,
			"thumbs": [],
			"type": "file"
		}`)); err != nil {
			return err
		}

		// add field
		if err := collection.Fields.AddMarshaledJSONAt(7, []byte(`{
			"hidden": false,
			"id": "bool2058414169",
			"name": "visible",
			"presentable": false,
			"required": false,
			"system": false,
			"type": "bool"
		}`)); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_4092854851")
		if err != nil {
			return err
		}

		// remove field
		collection.Fields.RemoveById("text2560465762")

		// remove field
		collection.Fields.RemoveById("text1843675174")

		// remove field
		collection.Fields.RemoveById("number3402113753")

		// remove field
		collection.Fields.RemoveById("text105650625")

		// remove field
		collection.Fields.RemoveById("file3309110367")

		// remove field
		collection.Fields.RemoveById("bool2058414169")

		return app.Save(collection)
	})
}
