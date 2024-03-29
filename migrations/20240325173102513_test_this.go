package migrations

import (
	"database/sql"

	"github.com/quangdinh/pg-migration/migrator"
)

type migration_20240325173102513 struct{}

func (m *migration_20240325173102513) Up(tx *sql.Tx) error {
	_, err := tx.Exec("CREATE TABLE mig01()")
	return err
}

func (m *migration_20240325173102513) Down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE mig01")
	return err
}

func init() {
	migrator.MustRegister(&migration_20240325173102513{})
}
