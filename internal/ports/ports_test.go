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
		{"mixed with duplicates", "22,20-23,22", []int{20, 21, 22, 23}, false},
		{"whitespace", " 22 , 80 ", []int{22, 80}, false},
		{"empty", "", nil, true},
		{"empty entry", "22,,80", nil, true},
		{"not a number", "http", nil, true},
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
