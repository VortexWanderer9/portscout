// Package ports parses port specifications such as "22,80,8000-8100".
package ports

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	minPort = 1
	maxPort = 65535
)

var webPorts = []int{80, 443, 8080, 8443}

var databasePorts = []int{1433, 1521, 3306, 5432, 6379, 27017}

var mailPorts = []int{25, 110, 143, 465, 587, 993, 995}

var remotePorts = []int{22, 23, 3389, 5900}

var dnsPorts = []int{53, 853}

// Parse turns a comma-separated list of ports and ranges into a sorted,
// de-duplicated slice of port numbers.
func Parse(spec string) ([]int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, errors.New("empty port specification")
	}
	seen := make(map[int]struct{})
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errors.New("empty entry in port specification")
		}
		if strings.EqualFold(part, "all") {
			for p := minPort; p <= maxPort; p++ {
				seen[p] = struct{}{}
			}
			continue
		}
		if strings.EqualFold(part, "web") {
			for _, p := range webPorts {
				seen[p] = struct{}{}
			}
			continue
		}
		if strings.EqualFold(part, "database") {
			for _, p := range databasePorts {
				seen[p] = struct{}{}
			}
			continue
		}
		if strings.EqualFold(part, "mail") {
			for _, p := range mailPorts {
				seen[p] = struct{}{}
			}
			continue
		}
		if strings.EqualFold(part, "remote") {
			for _, p := range remotePorts {
				seen[p] = struct{}{}
			}
			continue
		}
		if strings.EqualFold(part, "dns") {
			for _, p := range dnsPorts {
				seen[p] = struct{}{}
			}
			continue
		}

		lo, hi, err := parseEntry(part)
		if err != nil {
			return nil, err
		}
		for p := lo; p <= hi; p++ {
			seen[p] = struct{}{}
		}
	}

	out := make([]int, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Ints(out)
	return out, nil
}

// Exclude removes excluded ports from included while preserving sort order.
func Exclude(included, excluded []int) []int {
	blocked := make(map[int]struct{}, len(excluded))
	for _, p := range excluded {
		blocked[p] = struct{}{}
	}
	out := make([]int, 0, len(included))
	for _, p := range included {
		if _, ok := blocked[p]; !ok {
			out = append(out, p)
		}
	}
	return out
}

func parseEntry(s string) (lo, hi int, err error) {
	if a, b, isRange := strings.Cut(s, "-"); isRange {
		if lo, err = parsePort(a); err != nil {
			return 0, 0, err
		}
		if hi, err = parsePort(b); err != nil {
			return 0, 0, err
		}
		if lo > hi {
			return 0, 0, fmt.Errorf("invalid range %q: start is greater than end", s)
		}
		return lo, hi, nil
	}

	lo, err = parsePort(s)
	return lo, lo, err
}

func parsePort(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("invalid port %q", s)
	}
	if n < minPort || n > maxPort {
		return 0, fmt.Errorf("port %d out of range (%d-%d)", n, minPort, maxPort)
	}
	return n, nil
}
