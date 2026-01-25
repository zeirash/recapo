package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/logger"
)

// Tx is an interface that wraps sql.Tx methods used by services.
// This allows for mocking in tests.
type Tx interface {
	Commit() error
	Rollback() error
	QueryRow(query string, args ...any) *sql.Row
}

// DB is an interface that wraps the database methods used by services.
// This allows for mocking in tests.
type DB interface {
	Begin() (Tx, error)
	Close() error
}

// dbWrapper wraps *sql.DB to implement the DB interface.
type dbWrapper struct {
	db *sql.DB
}

func (w *dbWrapper) Begin() (Tx, error) {
	return w.db.Begin()
}

func (w *dbWrapper) Close() error {
	return w.db.Close()
}

var db *sql.DB

type DbConfig struct {
	User     string `required:"true" split_words:"true"`
	Password string `required:"true" split_words:"true"`
	Host     string `required:"true" split_words:"true"`
	Port     int    `required:"true" split_words:"true"`
	Name     string `required:"true" split_words:"true"`

	disableSSL bool
}

func url(dbConfig DbConfig) string {
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
	)

	if dbConfig.disableSSL {
		url = fmt.Sprintf("%s?sslmode=disable", url)
	}

	return url
}

func connect(conn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// InitDB initializes the database connection pool. Should be called once at startup.
func InitDB() {
	cfg := config.GetConfig()
	conn := url(DbConfig{
		User:     cfg.DbUsername,
		Password: cfg.DbPassword,
		Host:     cfg.DbHost,
		Port:     cfg.DbPort,
		Name:     cfg.DbName,

		disableSSL: true,
	})

	var err error
	db, err = connect(conn)
	if err != nil {
		logger.WithError(err).Fatal("failed to connect to database")
	}

	logger.Info("Database connection established")
}

// GetDB returns the database connection pool.
func GetDB() *sql.DB {
	if db == nil {
		logger.Fatal("database not initialized - call InitDB() first")
	}
	return db
}

// GetDBWrapper returns a DB interface wrapper around the database connection.
// This is used by services that need a mockable database interface.
func GetDBWrapper() DB {
	if db == nil {
		logger.Fatal("database not initialized - call InitDB() first")
	}
	return &dbWrapper{db: db}
}

// CloseDB closes the database connection. Should be called on shutdown.
func CloseDB() {
	if db != nil {
		db.Close()
	}
}
