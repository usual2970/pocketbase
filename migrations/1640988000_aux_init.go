package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.SystemMigrations.Add(&core.Migration{
		Up: func(txApp core.App) error {
			sql, err := txApp.AuxDBAdapter().CreateTableQuery(core.Table{
				Table: "_logs",
				Columns: []core.Field{
					&core.TextField{
						Name:       "id",
						PrimaryKey: true,
					},
					&core.TextField{
						Name: "level",
					},
					&core.TextField{
						Name: "message",
					},
					&core.JSONField{
						Name: "data",
					},
					&core.TextField{
						Name: "created",
					},
				},
				Indexes: []core.Index{
					{Name: "idx_logs_level", Columns: []string{"level"}},
					{Name: "idx_logs_message", Columns: []string{"message"}},
					{Name: "idx_logs_created_hour", Columns: []string{"created"}},
				},
			})

			if err != nil {
				return fmt.Errorf("create table query error: %w", err)
			}
			_, execErr := txApp.AuxDB().NewQuery(sql).Execute()

			return execErr
		},
		Down: func(txApp core.App) error {
			_, err := txApp.AuxDB().DropTable("_logs").Execute()
			return err
		},
		ReapplyCondition: func(txApp core.App, runner *core.MigrationsRunner, fileName string) (bool, error) {
			// reapply only if the _logs table doesn't exist
			exists := txApp.AuxHasTable("_logs")
			return !exists, nil
		},
	})
}
