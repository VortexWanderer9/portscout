package ports

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    []int
		wantErr bool
	}{
		{"single", "80", []int{80}, false},
		{"list", "80,22,443", []int{22, 80, 443}, false},
		{"range", "20-23", []int{20, 21, 22, 23}, false},
		{"web preset", "web", []int{80, 443, 8080, 8443}, false},
		{"web preset mixed", "22,web,443", []int{22, 80, 443, 8080, 8443}, false},
		{"database preset", "database", []int{1433, 1521, 3306, 5432, 6379, 27017}, false},
		{"mail preset", "mail", []int{25, 110, 143, 465, 587, 993, 995}, false},
		{"remote preset", "remote", []int{22, 23, 3389, 5900}, false},
		{"mixed with duplicates", "22,20-23,22", []int{20, 21, 22, 23}, false},
		{"whitespace", " 22 , 80 ", []int{22, 80}, false},
		{"empty", "", nil, true},
		{"empty entry", "22,,80", nil, true},
		{"not a number", "http", nil, true},
		{"missing range start", "-80", nil, true},
		{"missing range end", "80-", nil, true},
		{"multiple range separators", "20-22-24", nil, true},
		{"zero", "0", nil, true},
		{"too large", "70000", nil, true},
		{"reversed range", "100-50", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.spec, err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Parse(%q) = %v, want %v", tt.spec, got, tt.want)
			}
		})
	}
}

func TestParseAll(t *testing.T) {
	got, err := Parse("ALL")
	if err != nil {
		t.Fatalf("Parse(all): %v", err)
	}
	if len(got) != 65535 || got[0] != 1 || got[len(got)-1] != 65535 {
		t.Fatalf("Parse(all) returned unexpected range: len=%d first=%d last=%d", len(got), got[0], got[len(got)-1])
	}
}

func TestParseAllCanBeCombined(t *testing.T) {
	got, err := Parse("all,web")
	if err != nil {
		t.Fatalf("Parse(all,web): %v", err)
	}
	if len(got) != 65535 || got[0] != 1 || got[len(got)-1] != 65535 {
		t.Fatalf("Parse(all,web) returned unexpected range: len=%d first=%d last=%d", len(got), got[0], got[len(got)-1])
	}
}
