package domain

import "testing"

func TestParseMinorUnits(t *testing.T) {
	cases := []struct {
		text, currency string
		want           int64
		wantErr        bool
	}{{"0", "USD", 0, false}, {"12.3", "USD", 1230, false}, {"92233720368547758.08", "USD", 0, true}, {"1.001", "USD", 0, true}, {"01.00", "USD", 0, true}, {"-1", "USD", 0, true}, {"1", "EUR", 0, true}}
	for _, tc := range cases {
		got, err := ParseMinorUnits(tc.text, tc.currency)
		if (err != nil) != tc.wantErr || err == nil && got != tc.want {
			t.Errorf("ParseMinorUnits(%q,%q)=(%d,%v)", tc.text, tc.currency, got, err)
		}
	}
}
func FuzzParseMinorUnits(f *testing.F) {
	f.Add("1.23", "USD")
	f.Add("-4", "USD")
	f.Fuzz(func(t *testing.T, s, c string) {
		v, e := ParseMinorUnits(s, c)
		if e == nil && (c != "USD" || v < 0) {
			t.Fatalf("invalid accepted input %q %q => %d", s, c, v)
		}
	})
}
