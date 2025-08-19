package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func dataHandler() http.HandlerFunc {
	if strings.TrimSpace(os.Getenv("QUERY")) != "" {
		return queryHandler(connectDB(os.Getenv("DSN")))
	}

	return inputHandler()
}

func queryHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().AddDate(0, -1, 0)
		end := time.Now()
		if s := r.URL.Query().Get("start"); s != "" {
			t, err := time.Parse(time.DateOnly, s)
			if err == nil {
				start = t
			}
		}
		if s := r.URL.Query().Get("end"); s != "" {
			t, err := time.Parse(time.DateOnly, s)
			if err == nil {
				end = t
			}
		}

		results, err := getResults(r.Context(), db, start, end)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]any{
			"title":   "Report from query",
			"results": results,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func inputHandler() http.HandlerFunc {
	var (
		err     error
		input   = os.Stdin
		results = []timeSeriesData{}
	)

	if len(os.Args) > 1 {
		input, err = os.Open(filepath.Clean(os.Args[1]))
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}

	go func() {
		if err := processInput(input, &results); err != nil {
			slog.Error(err.Error())
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"title":   "Report from input",
			"results": results,
		}
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error(err.Error())
		}
	}
}

func processInput(reader io.Reader, results *[]timeSeriesData) error {
	fields := map[string]int{}
	for i, name := range lineRegex.SubexpNames() {
		fields[name] = i
	}

	buf := bufio.NewReader(reader)

scan:
	for bs, _, err := buf.ReadLine(); err != io.EOF; bs, _, err = buf.ReadLine() {
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}

		if len(bs) == 0 {
			continue
		}

		ts, err := processLine(bs, fields)
		if err != nil {
			return err
		}

		for i := range *results {
			if (*results)[i].Time.Equal(ts.Time) {
				(*results)[i].Value += ts.Value
				continue scan
			}
		}
		*results = append(*results, ts)
	}

	return nil
}

var (
	dtLogFormat = os.Getenv("DT_LOG_FORMAT")
	lineRegex   = regexp.MustCompile(os.Getenv("LOG_REGEX"))
)

func processLine(bs []byte, fields map[string]int) (timeSeriesData, error) {
	parts := lineRegex.FindStringSubmatch(string(bs))
	if len(parts) != 2 {
		return timeSeriesData{}, fmt.Errorf("invalid line format: %q", string(bs))
	}

	t, err := time.Parse(dtLogFormat, string(parts[fields["time"]]))
	if err != nil {
		return timeSeriesData{}, fmt.Errorf("invalid time format: %q", string(parts[fields["time"]]))
	}

	if t.Year() == 0 {
		t = t.AddDate(time.Now().Year(), 0, 0)
	}

	attrs := map[string]any{}
	for attr, pos := range fields {
		if attr == "time" || attr == "" {
			continue
		}
		attrs[attr] = parts[pos]
	}

	return timeSeriesData{Time: t, Value: 1, Attrs: attrs}, nil
}
