package migrator

import (
	"database/sql"
)

type Migration interface {
	Up(*sql.Tx) error
	Down(*sql.Tx) error
}
