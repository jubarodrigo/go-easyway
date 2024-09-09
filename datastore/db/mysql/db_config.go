package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"template/config"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

const (
	HostRead  = "HOST_READ"
	HostWrite = "HOST_WRITE"
)

type DBConfig struct {
	SkipMigrations bool `default:"false"`
	Settings       config.Settings
}

var connectTimeout = 15 * time.Second

type Config struct {
	Host            string        `json:"db_host"`
	HostRead        string        `json:"db_host_read"`
	Port            string        `json:"db_port" default:"3306"`
	Username        string        `json:"db_user" required:"true"`
	Password        string        `json:"db_password"`
	Database        string        `json:"db_database" required:"true"`
	ConnectTimeout  time.Duration `json:"db_connect_timeout" default:"2s"`
	MaxConnLifetime time.Duration `json:"db_max_conn_life_time" default:"1h"`
	MaxConns        int           `json:"db_max_conns" default:"5"`
	MinConns        int           `json:"db_min_conns" default:"5"`
	RefreshPassword bool          `json:"db_refresh_password" default:"false"`
	TZ              *string       `json:"db_time_zone"`
}

func (cfg DBConfig) EffectivePassword(ctx context.Context, awsConfig aws.Config) (string, error) {
	//TODO Rod - I believe this is never true because the calling method already checks if the password is empty?
	if cfg.Settings.Password != "" {
		return cfg.Settings.Password, nil
	}

	return auth.BuildAuthToken(ctx,
		fmt.Sprintf("%s:%s", cfg.Settings.Host, cfg.Settings.Port),
		awsConfig.Region, cfg.Settings.Username, awsConfig.Credentials,
	)
}

func (cfg DBConfig) ConnectionString(ctx context.Context, awsConfig aws.Config, hostType string) (string, error) {

	password := cfg.Settings.Password
	if len(password) == 0 {
		var err error
		password, err = cfg.EffectivePassword(ctx, awsConfig)
		if err != nil {
			return "", fmt.Errorf("could not create authentication token: %w", err)
		}
	}

	// host := switchDBHost(cfg.Settings, hostType)
	cfg.Settings.ConnectTimeout = connectTimeout

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%ds&allowCleartextPasswords=true", cfg.Settings.Username, password, cfg.Settings.Host, cfg.Settings.Port, cfg.Settings.Database,
		int(cfg.Settings.ConnectTimeout.Seconds()))

	return connString, nil
}

// DB contains connection details
type DB struct {
	Pool     *sql.DB
	PoolRead *sql.DB
	cfg      config.Settings
}

// NewDB creates a pooldb to initialize a database connection and run migrations
func NewDB(ctx context.Context, cfg DBConfig, awsConfig aws.Config) (DB, error) {

	log.Info().Msg(fmt.Sprintf("Creating new db connection: host: %s port: %s", cfg.Settings.Host, cfg.Settings.Port))
	dbPool, err := newPool(ctx, cfg, awsConfig, HostWrite)
	if err != nil {
		return DB{}, err
	}

	dbPoolRead, err := newPool(ctx, cfg, awsConfig, HostRead)
	if err != nil {
		return DB{}, err
	}

	return DB{
		Pool:     dbPool,
		PoolRead: dbPoolRead,
		cfg:      cfg.Settings,
	}, nil
}

func newPool(ctx context.Context, cfg DBConfig, awsConfig aws.Config, hostType string) (*sql.DB, error) {
	connStr, err := cfg.ConnectionString(ctx, awsConfig, hostType)
	if err != nil {
		log.Error().Err(err).Msg("failed to create connection string")
		return nil, err
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Error().Err(err).Msg("failed to open db connection")
		return nil, err
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(1)

	if err := db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("failed to ping db")
		return nil, err
	}

	return db, nil
}

// TODO: Implement this function when se have two different hosts
// func switchDBHost(cfg config.Settings, hostType string) string {
//	var host string
//
//	switch hostType {
//	case HostRead:
//		host = cfg.HostRead
//	default:
//		host = cfg.Host
//	}
//
//	return host
// }

func MustSetupDB(ctx context.Context, awscfg aws.Config, dbConfig DBConfig) DB {
	dbConn, err := NewDB(ctx, dbConfig, awscfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup db")
	}

	defaultConfig, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load aws config")
	}

	// Migrate the database
	log.Info().Msg("Migrating database...")
	err = Migrate(ctx, dbConfig, defaultConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to migrate db")
	}

	log.Info().Msg("DB setup successfully")

	return dbConn
}
