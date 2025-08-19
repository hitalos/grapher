package main

import (
	"bufio"
	"cmp"
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

		bs, err := json.Marshal(data)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(bs)
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

	trunc, err = time.ParseDuration(cmp.Or(os.Getenv("TIME_TRUNC"), "1m"))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	go func() {
		if err := processInput(input, &results); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"title":   "Report from input",
			"results": results,
		}

		bs, err := json.Marshal(data)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(bs)
	}
}

func processInput(reader io.Reader, results *[]timeSeriesData) error {
	fields := map[string]int{}
	for i, name := range lineRegex.SubexpNames() {
		fields[name] = i
	}

	var (
		buf = bufio.NewReader(reader)
		err error
		bs  []byte
	)

scan:
	for {
		bs, _, err = buf.ReadLine()
		if err != nil && err != io.EOF {
			break
		}

		if len(bs) == 0 {
			time.Sleep(1 * time.Second)
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

	return err
}

var (
	dtLogFormat = os.Getenv("DT_LOG_FORMAT")
	lineRegex   = regexp.MustCompile(os.Getenv("LOG_REGEX"))
	trunc       = time.Hour
)

func processLine(bs []byte, fields map[string]int) (timeSeriesData, error) {
	parts := lineRegex.FindStringSubmatch(string(bs))
	if len(parts) != 2 {
		return timeSeriesData{}, fmt.Errorf("invalid line format: %q", string(bs))
	}

	t, err := time.Parse(dtLogFormat, string(parts[fields["time"]]))
	if err != nil {
		return timeSeriesData{}, fmt.Errorf("invalid time format: '%s'", string(parts[fields["time"]]))
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

	return timeSeriesData{Time: t.Truncate(trunc), Value: 1, Attrs: attrs}, nil
}
