package migrator

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

var migrationsKey = []string{}
var migrations = map[string]Migration{}

const tableName = "_pgMigrationTable"

var (
	ErrUnresolvedVersion = errors.New("unresolved version")
)

func MustRegister(m Migration) {
	_, fpath, _, ok := runtime.Caller(1)
	if !ok {
		panic("Unable to get filename")
	}
	filename := filepath.Base(fpath)
	parts := strings.Split(filename, "_")
	version := parts[0]
	migrationsKey = append(migrationsKey, version)
	slices.Sort(migrationsKey)
	migrations[version] = m
}

func GetMigrations() []Migration {
	results := make([]Migration, len(migrationsKey))
	for i, k := range migrationsKey {
		results[i] = migrations[k]
	}

	return results
}

func Up(db *sql.DB) error {
	if len(migrationsKey) == 0 {
		return nil
	}
	version, err := GetCurrentVersion(db)
	if err != nil {
		return err
	}
	if version == "" {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		err = runUp(tx, migrations[migrationsKey[0]], migrationsKey[0])
		if err != nil {
			return errors.Join(err, tx.Rollback())
		}
		return tx.Commit()
	}

	for i, v := range migrationsKey {
		if v == version {
			if i+1 >= len(migrationsKey) {
				return nil
			}
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			err = runUp(tx, migrations[migrationsKey[i+1]], migrationsKey[i+1])
			if err != nil {
				return errors.Join(err, tx.Rollback())
			}
			return tx.Commit()
		}
	}
	return errors.Join(ErrUnresolvedVersion, fmt.Errorf("unknown version: %s", version))
}

func Down(db *sql.DB) error {
	if len(migrationsKey) == 0 {
		return nil
	}
	version, err := GetCurrentVersion(db)
	if err != nil {
		return err
	}
	if version == "" {
		return nil
	}

	for i, v := range migrationsKey {
		if v == version {

			tx, err := db.Begin()
			if err != nil {
				return err
			}
			newVersion := ""
			if i > 0 {
				newVersion = migrationsKey[i-1]
			}

			err = runDown(tx, migrations[migrationsKey[i]], newVersion)
			if err != nil {
				return errors.Join(err, tx.Rollback())
			}
			return tx.Commit()
		}
	}
	return errors.Join(ErrUnresolvedVersion, fmt.Errorf("unknown version: %s", version))
}

func Run(db *sql.DB) error {
	if len(migrationsKey) == 0 {
		return nil
	}
	version, err := GetCurrentVersion(db)
	if err != nil {
		return err
	}

	if version == "" {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		err = runMigrationsFromIndex(tx, 0)
		if err != nil {
			return errors.Join(err, tx.Rollback())
		}
		return tx.Commit()
	}

	for i, v := range migrationsKey {
		if v == version {
			if i == len(migrations)-1 {
				return nil
			}
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			err = runMigrationsFromIndex(tx, i+1)
			if err != nil {
				return errors.Join(err, tx.Rollback())
			}
			return tx.Commit()
		}
	}

	return errors.Join(ErrUnresolvedVersion, fmt.Errorf("unknown version: %s", version))
}

func runUp(tx *sql.Tx, m Migration, v string) error {
	err := m.Up(tx)
	if err != nil {
		return err
	}
	return setVersion(tx, v)
}

func runDown(tx *sql.Tx, m Migration, v string) error {
	err := m.Down(tx)
	if err != nil {
		return err
	}
	return setVersion(tx, v)
}

func setVersion(tx *sql.Tx, v string) error {
	_, err := tx.Exec(fmt.Sprintf("INSERT INTO %s (id, version) VALUES (1, $1) ON CONFLICT (id) DO UPDATE SET version = $1", tableName), v)
	return err
}

func runMigrationsFromIndex(tx *sql.Tx, i int) error {
	for ; i < len(migrationsKey); i++ {
		m := migrations[migrationsKey[i]]
		err := m.Up(tx)
		if err != nil {
			return err
		}
	}
	v := migrationsKey[len(migrationsKey)-1]
	return setVersion(tx, v)
}

func GetCurrentVersion(db *sql.DB) (string, error) {
	err := createMigrationTableIfNeeded(db)
	if err != nil {
		return "", err
	}

	rows := db.QueryRow(fmt.Sprintf("SELECT version FROM %s WHERE id = 1", tableName))
	var version string
	err = rows.Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return version, err
}

func createMigrationTableIfNeeded(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS %s (
      id int NOT NULL,
      version varchar(30) NOT NULL,
      PRIMARY KEY (id)
    )
`, tableName))
	return err
}
