// Package scanner implements a concurrent TCP connect scanner.
package scanner

import (
	"context"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Result describes the outcome of probing a single port.
type Result struct {
	Port    int           `json:"port"`
	Open    bool          `json:"open"`
	Service string        `json:"service,omitempty"`
	Banner  string        `json:"banner,omitempty"`
	Latency time.Duration `json:"latency_ns"`
}

// Options configures a scan.
type Options struct {
	Host        string
	Ports       []int
	Timeout     time.Duration // per-connection timeout
	Workers     int           // number of concurrent probes
	GrabBanners bool          // read a short banner from open ports
}

// Scan probes every port in opts.Ports and returns only the open ones,
// sorted by port number. It stops early if ctx is cancelled.
func Scan(ctx context.Context, opts Options) []Result {
	workers := workerCount(opts.Workers, len(opts.Ports))

	jobs := make(chan int)
	results := make(chan Result)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				if r := probe(ctx, opts, port); r.Open {
					select {
					case results <- r:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for _, p := range opts.Ports {
			select {
			case jobs <- p:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var open []Result
	for r := range results {
		open = append(open, r)
	}
	sort.Slice(open, func(i, j int) bool { return open[i].Port < open[j].Port })
	return open
}

func workerCount(requested, jobs int) int {
	if requested < 1 {
		return 1
	}
	if jobs > 0 && requested > jobs {
		return jobs
	}
	return requested
}

func probe(ctx context.Context, opts Options, port int) Result {
	res := Result{Port: port}
	addr := net.JoinHostPort(opts.Host, strconv.Itoa(port))

	dialer := net.Dialer{Timeout: opts.Timeout}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return res
	}
	defer conn.Close()

	res.Open = true
	res.Latency = time.Since(start)
	res.Service = ServiceName(port)

	if opts.GrabBanners {
		res.Banner = grabBanner(conn, opts.Timeout)
	}
	return res
}

func grabBanner(conn net.Conn, timeout time.Duration) string {
	wait := timeout / 2
	if wait > time.Second {
		wait = time.Second
	}
	_ = conn.SetReadDeadline(time.Now().Add(wait))

	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	return sanitize(string(buf[:n]))
}

// sanitize keeps the first line of a banner and strips non-printable bytes.
func sanitize(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}
