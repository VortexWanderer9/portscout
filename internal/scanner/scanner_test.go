package scanner

import (
	"context"
	"net"
	"testing"
	"time"
)

// listen starts a TCP listener on a random local port. If banner is non-empty,
// it is written to every accepted connection.
func listen(t *testing.T, banner string) (port int) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			if banner != "" {
				_, _ = c.Write([]byte(banner))
			}
			c.Close()
		}
	}()

	return ln.Addr().(*net.TCPAddr).Port
}

// closedPort returns a port that nothing is listening on.
func closedPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestScanFindsOnlyOpenPorts(t *testing.T) {
	open1 := listen(t, "")
	open2 := listen(t, "")
	closed := closedPort(t)

	got := Scan(context.Background(), Options{
		Host:    "127.0.0.1",
		Ports:   []int{closed, open1, open2},
		Timeout: 500 * time.Millisecond,
		Workers: 4,
	})

	if len(got) != 2 {
		t.Fatalf("expected 2 open ports, got %d: %+v", len(got), got)
	}
	if got[0].Port > got[1].Port {
		t.Errorf("results not sorted: %+v", got)
	}
	for _, r := range got {
		if r.Port == closed {
			t.Errorf("closed port %d reported as open", closed)
		}
	}
}

func TestScanGrabsBanner(t *testing.T) {
	port := listen(t, "SSH-2.0-TestServer\r\nmore data\r\n")

	got := Scan(context.Background(), Options{
		Host:        "127.0.0.1",
		Ports:       []int{port},
		Timeout:     time.Second,
		Workers:     1,
		GrabBanners: true,
	})

	if len(got) != 1 {
		t.Fatalf("expected 1 result, got %d", len(got))
	}
	if want := "SSH-2.0-TestServer"; got[0].Banner != want {
		t.Errorf("banner = %q, want %q", got[0].Banner, want)
	}
}

func TestScanRespectsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ports := make([]int, 0, 1000)
	for p := 1; p <= 1000; p++ {
		ports = append(ports, p)
	}

	done := make(chan struct{})
	go func() {
		Scan(ctx, Options{Host: "127.0.0.1", Ports: ports, Timeout: time.Second, Workers: 8})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Scan did not return promptly after context cancellation")
	}
}

func TestSanitize(t *testing.T) {
	if got := sanitize("hello\x00\x01 world\nsecond line"); got != "hello world" {
		t.Errorf("sanitize = %q", got)
	}
}

func TestServiceName(t *testing.T) {
	for port, want := range map[int]string{
		22:    "ssh",
		389:   "ldap",
		2375:  "docker",
		6443:  "kubernetes",
		11211: "memcached",
	} {
		if got := ServiceName(port); got != want {
			t.Errorf("ServiceName(%d) = %q, want %q", port, got, want)
		}
	}
	if ServiceName(54321) != "" {
		t.Error("expected empty name for unknown port")
	}
}

func TestWorkerCount(t *testing.T) {
	tests := []struct {
		requested int
		jobs      int
		want      int
	}{
		{requested: 0, jobs: 5, want: 1},
		{requested: 2, jobs: 5, want: 2},
		{requested: 10, jobs: 2, want: 2},
		{requested: 10, jobs: 0, want: 10},
	}
	for _, tt := range tests {
		if got := workerCount(tt.requested, tt.jobs); got != tt.want {
			t.Errorf("workerCount(%d, %d) = %d, want %d", tt.requested, tt.jobs, got, tt.want)
		}
	}
}
