package database

import (
	"context"
	"fmt"
	"time"

	"github.com/black-coda/go-boilerplate/internal/config"
	loggerConfig "github.com/black-coda/go-boilerplate/internal/logger"
	pgxzero "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/newrelic/go-agent/v3/integrations/nrpgx5"
	"github.com/rs/zerolog"
)

const (
	DatabasePingTimeout = 10
)

type Database struct {
	Pool *pgxpool.Pool
	log  *zerolog.Logger
}

type multiTracer struct {
	tracers []any
}

// Parameter =>	Meaning
//mt *multiTracer	=> The receiver. It gives the method access to mt.tracers.
//ctx context.Context =>	Carries cancellation signals and request-specific values.
//conn *pgx.Conn =>	Points to the PostgreSQL connection that runs the query.
//data pgx.TraceQueryStartData =>	Contains information about the query that starts.

func (mt *multiTracer) TraceQueryStart(
	ctx context.Context,
	conn *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	for _, tracer := range mt.tracers {
		// Iterate through each registered tracer.
		// Using an inline interface type assertion, we check if the current tracer
		// implements the 'TraceQueryStartData' method. If it does, we safely invoke it;
		// if not, the tracer is skipped.
		if t, ok := tracer.(interface {
			TraceQueryStart(
				context.Context,
				*pgx.Conn,
				pgx.TraceQueryStartData,
			) context.Context
		}); ok {
			ctx = t.TraceQueryStart(ctx, conn, data)
		}
	}
	return ctx
}

func (mt *multiTracer) TraceQueryEnd(ctx context.Context,
	conn *pgx.Conn,
	data pgx.TraceQueryEndData,
) {
	for _, tracer := range mt.tracers {
		// Iterate through each registered tracer.
		// Using an inline interface type assertion, we check if the current tracer
		// implements the 'TraceQueryStartData' method. If it does, we safely invoke it;
		// if not, the tracer is skipped.
		if t, ok := tracer.(interface {
			TraceQueryEnd(
				context.Context,
				*pgx.Conn,
				pgx.TraceQueryEndData,
			) context.Context
		}); ok {
			t.TraceQueryEnd(ctx, conn, data)
		}
	}

}

func New(cfg *config.Config, logger *zerolog.Logger, loggerService *loggerConfig.LoggerService) (*Database, error) {

	// URL-encode the password
	// Issue:
	//Example
	//
	//Your password is my pass.
	//
	//QueryEscape changes it to my+pass.
	//pgx reads the DSN. In the password part, + means a plus sign.
	//pgx sends the password my+pass to the database.
	//The database says the password is wrong, because the real password is my pass.
	//
	// Former method:
	//hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))
	//encodedPassword := url.QueryEscape(cfg.Database.Password)

	//postgresql://username:password@host:port/database_name
	//dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
	//	cfg.Database.User,
	//	encodedPassword,
	//	hostPort,
	//	cfg.Database.Name,
	//	cfg.Database.SSLMode,
	//)
	dsn := DsnConfig(cfg)

	pgxPoolConfig, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
	}

	// Add New Relic PostgreSQL instrumentation
	if loggerService != nil && loggerService.GetApplication() != nil {
		pgxPoolConfig.ConnConfig.Tracer = nrpgx5.NewTracer()
	}

	if cfg.Primary.Env == "local" {
		globalLevel := logger.GetLevel()
		pgxLogger := loggerConfig.NewPgxLogger(globalLevel)
		// Chain tracers - New Relic first, then local logging
		if pgxPoolConfig.ConnConfig.Tracer != nil {
			// If New Relic tracer exists, create a multi-tracer
			localTracer := &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLevel)),
			}
			pgxPoolConfig.ConnConfig.Tracer = &multiTracer{
				tracers: []any{pgxPoolConfig.ConnConfig.Tracer, localTracer},
			}
		} else {
			pgxPoolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
				Logger:   pgxzero.NewLogger(pgxLogger),
				LogLevel: tracelog.LogLevel(loggerConfig.GetPgxTraceLogLevel(globalLevel)),
			}
		}
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	database := &Database{
		Pool: pool,
		log:  logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), DatabasePingTimeout*time.Second)
	defer cancel()
	if err = pool.Ping(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().Msg("connected to the database")

	return database, nil
}

func (db *Database) Close() error {
	db.log.Info().Msg("closing database connection pool")
	db.Pool.Close()
	return nil
}
