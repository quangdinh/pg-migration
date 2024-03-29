package main

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/quangdinh/pg-migration/migrations"
	"github.com/quangdinh/pg-migration/migrator"
	"github.com/stretchr/testify/assert"
)

func TestRunMigration(t *testing.T) {
	migrations := migrator.GetMigrations()
	assert.Len(t, migrations, 3)
	m := migrations[0]
	assert.Equal(t, "migration_20240325173102513", reflect.TypeOf(m).Elem().Name())
	m = migrations[1]
	assert.Equal(t, "migration_20240325174351002", reflect.TypeOf(m).Elem().Name())

	m = migrations[2]
	assert.Equal(t, "migration_20240325174354720", reflect.TypeOf(m).Elem().Name())
}

func clean(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS _pgMigrationTable")
	db.Exec("DROP TABLE IF EXISTS mig01")
	db.Exec("DROP TABLE IF EXISTS mig02")
	db.Exec("DROP TABLE IF EXISTS mig03")
}

func TestRun(t *testing.T) {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=postgres password=postgres dbname=migration sslmode=disable")
	assert.NoError(t, err)
	defer db.Close()
	clean(db)
	err = migrator.Run(db)
	assert.NoError(t, err)
}

func TestDownUp(t *testing.T) {
	db, err := sql.Open("pgx", "host=localhost port=5432 user=postgres password=postgres dbname=migration sslmode=disable")
	assert.NoError(t, err)
	defer db.Close()
	clean(db)
	err = migrator.Up(db)
	assert.NoError(t, err)

	var exists bool

	version, err := migrator.GetCurrentVersion(db)
	assert.NoError(t, err)
	assert.Equal(t, "20240325173102513", version)
	row := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'mig01') AS table_existence")
	err = row.Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)

	err = migrator.Up(db)
	assert.NoError(t, err)
	version, err = migrator.GetCurrentVersion(db)
	assert.NoError(t, err)
	assert.Equal(t, "20240325174351002", version)
	row = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'mig02') AS table_existence")
	err = row.Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists)

	err = migrator.Down(db)
	assert.NoError(t, err)
	version, err = migrator.GetCurrentVersion(db)
	assert.NoError(t, err)
	assert.Equal(t, "20240325173102513", version)
	row = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'mig02') AS table_existence")
	err = row.Scan(&exists)
	assert.NoError(t, err)
	assert.False(t, exists)
}
