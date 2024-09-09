package mysql

import (
	"context"
	"database/sql"
	"embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Migrate(ctx context.Context, cfg DBConfig, awsCfg aws.Config) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	connStr, err := cfg.ConnectionString(ctx, awsCfg, "")
	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.Up(db, "migrations", goose.WithAllowMissing())
}
