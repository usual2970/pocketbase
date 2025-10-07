package migrations

import (
	"github.com/pocketbase/pocketbase/core"
)

// note: this migration will be deleted in future version

func init() {
	core.SystemMigrations.Register(func(txApp core.App) error {
		sql, err := txApp.DBAdapter().CreateIndexQuery(core.Index{
			Name:      "idx__collections_type",
			TableName: "_collections",
			Columns:   []string{"type"},
			Unique:    false,
		})
		if err != nil {
			return err
		}
		_, err = txApp.DB().NewQuery(sql).Execute()
		if err != nil {
			return err
		}

		// reset mfas and otps delete rule
		collectionNames := []string{core.CollectionNameMFAs, core.CollectionNameOTPs}
		for _, name := range collectionNames {
			col, err := txApp.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}

			if col.DeleteRule != nil {
				col.DeleteRule = nil
				err = txApp.SaveNoValidate(col)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}, nil)
}
