// Package db provides an embedded filesystem containing all the database migrations
package db

import (
	"embed"
)

// Migrations contain an embedded filesystem with all the sql migration files
//
//go:embed migrations/*.sql
var Migrations embed.FS

// GooseMigrationsSQLite contain an embedded filesystem with all the goose migration files for sqlite
//
//go:embed migrations-goose-sqlite/*.sql
var GooseMigrationsSQLite embed.FS

// GooseMigrationsPG contain an embedded filesystem with all the goose migration files for postgres
//
//go:embed migrations-goose-postgres/*.sql
var GooseMigrationsPG embed.FS
