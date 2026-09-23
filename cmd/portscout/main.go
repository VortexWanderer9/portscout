// Command portscout is a fast, concurrent TCP port scanner.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/VortexWanderer9/portscout/internal/ports"
	"github.com/VortexWanderer9/portscout/internal/report"
	"github.com/VortexWanderer9/portscout/internal/scanner"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(runArgs(os.Args[1:], os.Stdout, os.Stderr))
}

func runArgs(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		switch args[0] {
		case "version":
			fmt.Fprintln(stdout, "portscout", version)
			return 0
		case "scan":
			args = args[1:]
		case "help":
			args = []string{"--help"}
		}
	}

	fs := flag.NewFlagSet("portscout scan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "portscout %s - concurrent TCP port scanner\n\n", version)
		fmt.Fprintln(stderr, "Usage:")
		fmt.Fprintln(stderr, "  portscout scan [flags] <host>")
		fmt.Fprintln(stderr, "  portscout version")
		fmt.Fprintln(stderr, "\nThe legacy form 'portscout [flags] <host>' remains supported.\n\nFlags:")
		fs.PrintDefaults()
		fmt.Fprintln(stderr, "\nOnly scan hosts you own or have permission to test.")
	}

	var (
		portSpec    string
		excludeSpec string
		timeout     time.Duration
		workers     int
		banners     bool
		format      string
		network     string
		asJSON      bool
		asCSV       bool
		quiet       bool
		showVersion bool
		help        bool
	)
	fs.StringVar(&portSpec, "p", "1-1024", "ports to scan, e.g. \"22,80,443\" or \"1-1024\"")
	fs.StringVar(&portSpec, "ports", "1-1024", "ports to scan")
	fs.StringVar(&excludeSpec, "exclude", "", "ports to skip")
	fs.DurationVar(&timeout, "t", 800*time.Millisecond, "timeout per connection")
	fs.DurationVar(&timeout, "timeout", 800*time.Millisecond, "timeout per connection")
	fs.IntVar(&workers, "w", 200, "number of concurrent workers")
	fs.IntVar(&workers, "workers", 200, "number of concurrent workers")
	fs.BoolVar(&banners, "b", false, "grab service banners from open ports")
	fs.BoolVar(&banners, "banners", false, "grab service banners from open ports")
	fs.StringVar(&format, "format", "table", "output format: table, json, csv, or ports")
	fs.StringVar(&network, "network", "tcp", "network: tcp, tcp4, or tcp6")
	fs.BoolVar(&asJSON, "json", false, "output results as JSON (legacy alias)")
	fs.BoolVar(&asCSV, "csv", false, "output results as CSV (legacy alias)")
	fs.BoolVar(&quiet, "quiet", false, "output only open port numbers (legacy alias)")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.BoolVar(&showVersion, "v", false, "print version and exit")
	fs.BoolVar(&help, "help", false, "print help and exit")
	fs.BoolVar(&help, "h", false, "print help and exit")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if help {
		fs.Usage()
		return 0
	}
	if showVersion {
		fmt.Fprintln(stdout, "portscout", version)
		return 0
	}
	if fs.NArg() != 1 {
		fs.Usage()
		return 2
	}
	if workers < 1 {
		fmt.Fprintln(stderr, "error: --workers must be at least 1")
		return 2
	}
	if timeout <= 0 {
		fmt.Fprintln(stderr, "error: --timeout must be greater than 0")
		return 2
	}

	formatSet := false
	fs.Visit(func(f *flag.Flag) {
		formatSet = formatSet || f.Name == "format"
	})
	if formatSet && boolCount(asJSON, asCSV, quiet) > 0 {
		fmt.Fprintln(stderr, "error: --format cannot be combined with --json, --csv, or --quiet")
		return 2
	}
	if !formatSet {
		switch {
		case asJSON:
			format = "json"
		case asCSV:
			format = "csv"
		case quiet:
			format = "ports"
		}
	}
	if !validFormat(format) {
		fmt.Fprintf(stderr, "error: unsupported format %q (use table, json, csv, or ports)\n", format)
		return 2
	}
	if !validNetwork(network) {
		fmt.Fprintf(stderr, "error: unsupported network %q (use tcp, tcp4, or tcp6)\n", network)
		return 2
	}
	if boolCount(asJSON, asCSV, quiet) > 1 {
		fmt.Fprintln(stderr, "error: only one of --json, --csv, or --quiet may be used")
		return 2
	}

	portList, err := ports.Parse(portSpec)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	if excludeSpec != "" {
		excluded, err := ports.Parse(excludeSpec)
		if err != nil {
			fmt.Fprintf(stderr, "error: invalid --exclude value: %v\n", err)
			return 2
		}
		portList = ports.Exclude(portList, excluded)
		if len(portList) == 0 {
			fmt.Fprintln(stderr, "error: --exclude removed every selected port")
			return 2
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	host := fs.Arg(0)
	start := time.Now()
	open := scanner.Scan(ctx, scanner.Options{
		Host:        host,
		Network:     network,
		Ports:       portList,
		Timeout:     timeout,
		Workers:     workers,
		GrabBanners: banners,
	})

	summary := report.Summary{
		Host:     host,
		Scanned:  len(portList),
		Duration: time.Since(start),
		Open:     open,
	}

	switch format {
	case "ports":
		err = report.Ports(stdout, open)
	case "json":
		err = report.JSON(stdout, summary)
	case "csv":
		err = report.CSV(stdout, open)
	default:
		err = report.Table(stdout, summary)
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if ctx.Err() != nil {
		fmt.Fprintln(stderr, "scan interrupted; results are partial")
		return 130
	}
	return 0
}

func validNetwork(network string) bool {
	return network == "tcp" || network == "tcp4" || network == "tcp6"
}

func validFormat(format string) bool {
	switch format {
	case "table", "json", "csv", "ports":
		return true
	default:
		return false
	}
}

func boolCount(values ...bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}
