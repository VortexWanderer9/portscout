// Command portscout is a fast, concurrent TCP port scanner.
package main

import (
	"context"
	"flag"
	"fmt"
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
	os.Exit(run())
}

func run() int {
	var (
		portSpec    = flag.String("p", "1-1024", "ports to scan, e.g. \"22,80,443\" or \"1-1024\"")
		timeout     = flag.Duration("t", 800*time.Millisecond, "timeout per connection")
		workers     = flag.Int("w", 200, "number of concurrent workers")
		banners     = flag.Bool("b", false, "grab service banners from open ports")
		asJSON      = flag.Bool("json", false, "output results as JSON")
		quiet       = flag.Bool("quiet", false, "output only open port numbers")
		showVersion = flag.Bool("version", false, "print version and exit")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "portscout %s - concurrent TCP port scanner\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n  portscout [flags] <host>\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nOnly scan hosts you own or have permission to test.\n")
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("portscout", version)
		return 0
	}
	if flag.NArg() != 1 {
		flag.Usage()
		return 2
	}
	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "error: -w must be at least 1")
		return 2
	}
	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "error: -t must be greater than 0")
		return 2
	}
	if *asJSON && *quiet {
		fmt.Fprintln(os.Stderr, "error: -json and -quiet cannot be used together")
		return 2
	}

	portList, err := ports.Parse(*portSpec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 2
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	host := flag.Arg(0)
	start := time.Now()
	open := scanner.Scan(ctx, scanner.Options{
		Host:        host,
		Ports:       portList,
		Timeout:     *timeout,
		Workers:     *workers,
		GrabBanners: *banners,
	})

	summary := report.Summary{
		Host:     host,
		Scanned:  len(portList),
		Duration: time.Since(start),
		Open:     open,
	}

	if *quiet {
		err = report.Ports(os.Stdout, open)
	} else if *asJSON {
		err = report.JSON(os.Stdout, summary)
	} else {
		err = report.Table(os.Stdout, summary)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	if ctx.Err() != nil {
		fmt.Fprintln(os.Stderr, "scan interrupted; results are partial")
		return 130
	}
	return 0
}
