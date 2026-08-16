package app

import "testing"

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{name: "zero", amount: 0, want: "$0.00"},
		{name: "thousands", amount: 1234.5, want: "$1,234.50"},
		{name: "negative hundreds", amount: -123, want: "-$123.00"},
		{name: "negative thousands", amount: -123456, want: "-$123,456.00"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := formatMoney(test.amount); got != test.want {
				t.Fatalf("formatMoney(%v) = %q, want %q", test.amount, got, test.want)
			}
		})
	}
}
