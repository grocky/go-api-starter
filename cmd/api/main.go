package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"

	"github.com/grocky/go-api-starter/cmd/api/app"
	"github.com/grocky/go-api-starter/cmd/api/server"
	"github.com/grocky/go-api-starter/internal/log"
	"github.com/grocky/go-api-starter/internal/mysql"
)

const appName = "go-api-starter"

func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
	)

	logger := log.DefaultLogger()
	ctx = log.WithLogger(ctx, logger)

	defer func() {
		done()
		if r := recover(); r != nil {
			logger.Error("application panic", "panic", r)
			os.Exit(9)
		}
	}()

	err := run(ctx)
	done()

	if err != nil {
		logger.Error(err.Error(), "stack", debug.Stack())
		os.Exit(2)
	}
	logger.Warn("successful shutdown")
}

type config struct {
	httpPort int
	db       struct {
		host     string
		port     int
		user     string
		password string
	}
	version bool
}

func run(ctx context.Context) error {
	logger := log.FromContext(ctx).Named(appName)
	ctx = log.WithLogger(ctx, logger)

	var cfg config
	var err error

	cfg.httpPort = 3000
	if os.Getenv("APP_PORT") != "" {
		cfg.httpPort, err = strconv.Atoi(os.Getenv("APP_PORT"))
		if err != nil {
			return fmt.Errorf("APP_PORT must be an integer, %w", err)
		}
	}
	cfg.db.host = os.Getenv("DB_HOST")
	if cfg.db.port, err = strconv.Atoi(os.Getenv("DB_PORT")); err != nil {
		return fmt.Errorf("DB_PORT must be an integer, %w", err)
	}
	cfg.db.user = os.Getenv("DB_USER")
	cfg.db.password = os.Getenv("DB_PASS")

	var db *mysql.DB
	dbConfig := mysql.NewConfig(appName, cfg.db.host, cfg.db.port, cfg.db.user, cfg.db.password)
	if db, err = mysql.New(ctx, dbConfig); err != nil {
		logger.Error("unable to connect to mysql", "host", cfg.db.host, "port", cfg.db.port, "error", err)
		return fmt.Errorf("unable to connect to mysql")
	}
	defer func(db *mysql.DB) {
		if dbErr := db.Close(); err != nil {
			logger.Error("error while closing mysql connection", dbErr)
		}
	}(db)

	app := app.New(db)

	srv, err := server.New(cfg.httpPort)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}

	logger.Info("server listening", "port", cfg.httpPort)

	return srv.ServeHTTPHandler(ctx, app.Routes(ctx))
}
