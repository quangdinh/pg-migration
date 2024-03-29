package migrations

import (
	"database/sql"

	"github.com/quangdinh/pg-migration/migrator"
)

type migration_20240325174354720 struct{}

func (m *migration_20240325174354720) Up(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE mig03()")
	return err
}

func (m *migration_20240325174354720) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE mig03")
	return err
}

func init() {
	migrator.MustRegister(&migration_20240325174354720{})
}
