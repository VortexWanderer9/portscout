package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/VortexWanderer9/portscout/internal/scanner"
)

func sample() Summary {
	return Summary{
		Host:     "example.test",
		Scanned:  100,
		Duration: 2 * time.Second,
		Open: []scanner.Result{
			{Port: 22, Open: true, Service: "ssh", Banner: "SSH-2.0-OpenSSH", Latency: 3 * time.Millisecond},
		},
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	if err := Table(&buf, sample()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"example.test", "22/tcp", "ssh", "SSH-2.0-OpenSSH", "1 open, 100 scanned"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestTableNoResults(t *testing.T) {
	var buf bytes.Buffer
	s := sample()
	s.Open = nil
	if err := Table(&buf, s); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No open ports found") {
		t.Errorf("unexpected output: %s", buf.String())
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, sample()); err != nil {
		t.Fatal(err)
	}
	var got Summary
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Host != "example.test" || len(got.Open) != 1 || got.Open[0].Port != 22 {
		t.Errorf("round trip mismatch: %+v", got)
	}
}

func TestPorts(t *testing.T) {
	var buf bytes.Buffer
	if err := Ports(&buf, []scanner.Result{{Port: 22}, {Port: 443}}); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "22\n443\n"; got != want {
		t.Errorf("Ports() = %q, want %q", got, want)
	}
}

func TestCSV(t *testing.T) {
	var buf bytes.Buffer
	results := []scanner.Result{{Port: 443, Service: "https", Latency: time.Millisecond, Banner: "server, v1"}}
	if err := CSV(&buf, results); err != nil {
		t.Fatal(err)
	}
	if got, want := buf.String(), "port,state,service,latency,banner\n443,open,https,1ms,\"server, v1\"\n"; got != want {
		t.Errorf("CSV() = %q, want %q", got, want)
	}
}
