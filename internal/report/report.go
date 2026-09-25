// Package report renders scan results as a table or JSON.
package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/VortexWanderer9/portscout/internal/scanner"
)

// Summary carries scan metadata alongside the results.
type Summary struct {
	Host        string           `json:"host"`
	Hostname    string           `json:"hostname"`
	Scanned     int              `json:"ports_scanned"`
	Excluded    int              `json:"ports_excluded"`
	Duration    time.Duration    `json:"duration_ns"`
	Workers     int              `json:"workers"`
	Timeout     time.Duration    `json:"timeout_ns"`
	Network     string           `json:"network"`
	GrabBanners bool             `json:"grab_banners"`
	Open        []scanner.Result `json:"open"`
}

// JSON writes the summary as indented JSON.
func JSON(w io.Writer, s Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// Ports writes one open TCP port number per line for shell pipelines.
func Ports(w io.Writer, open []scanner.Result) error {
	for _, r := range open {
		if _, err := fmt.Fprintln(w, r.Port); err != nil {
			return err
		}
	}
	return nil
}

// CSV writes open ports with a header row for use in spreadsheets and scripts.
func CSV(w io.Writer, open []scanner.Result) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"port", "state", "service", "latency", "banner"}); err != nil {
		return err
	}
	for _, r := range open {
		if err := cw.Write([]string{
			fmt.Sprintf("%d", r.Port),
			"open",
			r.Service,
			r.Latency.Round(time.Microsecond).String(),
			r.Banner,
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// Table writes a human-readable table.
func Table(w io.Writer, s Summary) error {
	fmt.Fprintf(w, "Scan report for %s\n", s.Host)

	if len(s.Open) == 0 {
		fmt.Fprintf(w, "No open ports found (%d scanned in %s)\n", s.Scanned, s.Duration.Round(time.Millisecond))
		return nil
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PORT\tSTATE\tSERVICE\tLATENCY\tBANNER")
	for _, r := range s.Open {
		svc := r.Service
		if svc == "" {
			svc = "-"
		}
		fmt.Fprintf(tw, "%d/tcp\topen\t%s\t%s\t%s\n", r.Port, svc, r.Latency.Round(time.Microsecond), r.Banner)
	}
	if err := tw.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(w, "\n%d open, %d scanned in %s\n", len(s.Open), s.Scanned, s.Duration.Round(time.Millisecond))
	return nil
}
