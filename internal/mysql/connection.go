package mysql

import (
	"context"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/grocky/go-api-starter/internal/log"
	"net/url"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

const defaultTimeout = 3 * time.Second

type DB struct {
	*sqlx.DB
}

func New(ctx context.Context, options Config) (*DB, error) {
	logger := log.FromContext(ctx).Named("mysql")
	logger.Info("connecting to mysql", "dsn", options.DSN())

	db, err := sqlx.Connect(options.driver, options.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(options.MaxOpenConns)
	db.SetMaxIdleConns(options.MaxIdleConns)
	db.SetConnMaxIdleTime(options.MaxIdleTime)
	db.SetConnMaxLifetime(options.MaxLifetime)

	return &DB{db}, nil
}

type (
	// Config include common connection options
	Config struct {
		mysql.Config
		Username string
		Password string
		Host     string
		Port     int
		Database string
		driver   string

		Retries        int
		RetryDelay     time.Duration
		ConnectTimeout time.Duration
		MaxOpenConns   int
		MaxIdleConns   int
		MaxIdleTime    time.Duration
		MaxLifetime    time.Duration
	}
)

var dsnHostReplacer = regexp.MustCompile("//(.*)@(.*):(.*)/")

func (co *Config) DSN() string {
	u := url.URL{
		Host: fmt.Sprintf("%s:%d", co.Host, co.Port),
		User: url.UserPassword(co.Username, co.Password),
		Path: co.Database,
	}

	params := u.Query()
	params.Add("collation", "utf8_general_ci")
	params.Add("parseTime", "true")
	params.Add("loc", "UTC")

	u.RawQuery = params.Encode()

	return dsnHostReplacer.ReplaceAllString(u.String(), "$1@tcp($2:$3)/")
}

func NewConfig(schemaName, host string, port int, username, password string) Config {
	return Config{
		Config: mysql.Config{
			User:   username,
			Passwd: password,
			Net:    "tcp",
			Addr:   fmt.Sprintf("%s:%d", host, port),
			DBName: schemaName,
		},
		Username: username,
		Password: password,
		Host:     host,
		Port:     port,
		Database: schemaName,
		driver:   "mysql",

		Retries:        3,
		RetryDelay:     time.Second * 3,
		ConnectTimeout: time.Minute * 1,

		MaxOpenConns: 25,
		MaxIdleConns: 25,
		MaxIdleTime:  5 * time.Minute,
		MaxLifetime:  2 * time.Hour,
	}
}
