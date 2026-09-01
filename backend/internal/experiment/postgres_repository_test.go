package experiment

import (
	"bytes"
	"testing"
)

func TestNormalizeLegacyJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain JSON", input: `["MA","RSI"]`, want: `["MA","RSI"]`},
		{name: "legacy bytea text", input: `\x5b224d41225d`, want: `["MA"]`},
		{name: "malformed bytea text", input: `\xnot-hex`, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeLegacyJSON([]byte(test.input))
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && !bytes.Equal(got, []byte(test.want)) {
				t.Fatalf("value = %q, want %q", got, test.want)
			}
		})
	}
}
