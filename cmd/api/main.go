package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"greenlight.oskr.nl/internal/data"
	"greenlight.oskr.nl/internal/jsonlog"
	"greenlight.oskr.nl/internal/mailer"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// config struct
type config struct {
	port 	int
	env 	string // dev | staging | prod
	db struct {
		dsn 					string
		maxOpenConns 	int
		maxIdleConns 	int
		maxIdleTime 	string
	}
	limiter struct {
		rps 			float64
		burst 		int
		enabled 	bool
	}
	smtp struct {
		host 			string
		port 			int
		username 	string
		password 	string
		sender 		string
	}
	cors struct {
		trustedOrigins []string
	}
}


// application struct to hold depends for our HTTP handlers,
// helpers, and middleware.
type application struct {
	config 	config
	logger 	*jsonlog.Logger
	models 	data.Models
	mailer 	mailer.Mailer
	wg 			sync.WaitGroup
}

func main() {
	var cfg config

	// read and parse command line args
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (dev|staging|prod)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connection")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connection")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	
	// rate limit conf
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	
	// mailer conf
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "eb1586609f1089", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "6af0fce66ece34", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight HQ <ghq@greenlight.oskr.nl>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space seperated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	// init a new logger which writes to stdout with date and time
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("db connection pool established", nil)

	expvar.NewString("version").Set(version)

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() interface{} {
		return db.Stats()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Unix()
	}))

	// declare an instance of application struct which contains
	// our config and the logger
	app := &application {
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(
			cfg.smtp.host, cfg.smtp.port, 
			cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender,
		),
	}


	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}