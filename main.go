package main

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hitalos/grapher/public"
)

var (
	isDevMode = strings.ToLower(os.Getenv("ENV")) == "dev"
	logger    = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: isDevMode}))
)

type timeSeriesData struct {
	Time  time.Time      `json:"time"`
	Value int            `json:"value"`
	Attrs map[string]any `json:"attrs,omitempty"`
}

func main() {
	slog.SetDefault(logger)

	http.Handle("/", http.FileServer(public.FS))
	http.HandleFunc("/data", dataHandler())

	s := http.Server{
		Addr:         cmp.Or(os.Getenv("PORT"), ":6060"),
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  30 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      logMiddleware(gzipMiddleware(http.DefaultServeMux)),
	}
	slog.Info("Listening", "interface", s.Addr)
	dieOnErr(s.ListenAndServe())
}

func dieOnErr(err error) {
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func connectDB(dsn string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(dsn)
	dieOnErr(err)
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	cfg.ConnConfig.RuntimeParams["application_name"] = os.Args[0]

	db, err := pgxpool.NewWithConfig(context.Background(), cfg)
	dieOnErr(err)

	return db
}

func getResults(ctx context.Context, db *pgxpool.Pool, start, end time.Time) ([]timeSeriesData, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, os.Getenv("QUERY"), start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	switch {
	case fields[0].Name != "time":
		return nil, errors.New("first field returned by query must be 'time'")
	case fields[1].Name != "value":
		return nil, errors.New("second field returned by query must be 'value'")
	}

	return pgx.CollectRows(rows, pgx.RowToStructByNameLax[timeSeriesData])
}
