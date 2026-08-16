package stock

import (
	"strings"
	"testing"
)

func TestParseTSVAggregatesVestsAndSkipsDividends(t *testing.T) {
	input := `Award Date\tSymbol\tAward ID\tShare Type\tMarket Value\tDate Holding Period Met\tDeposit Date\tDate Acquired\tAcquisition Price\tShares\tAvailable to Sell
03/06/2024\tGOOG\tPRIVATE-ID-1\tRestricted Stock\t$455.45\tN/A\t07/29/2026\t07/25/2026\t$319.09\t2.021\t1.321
03/05/2025\tGOOG\tPRIVATE-ID-2\tRestricted Stock\t$453.38\tN/A\t07/29/2026\t07/25/2026\t$319.09\t2.011\t1.315
N/A\tGOOG\t--\tDividend Reinvestment\t$39.37\tN/A\t06/15/2026\t06/15/2026\t$368.84\t0.1142\t0.1142`

	input = strings.ReplaceAll(input, `\t`, "\t")
	vests, err := ParseTSV(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(vests) != 1 {
		t.Fatalf("got %d vest events, want 1", len(vests))
	}
	vest := vests[0]
	if vest.GrossSharesMicros != 4_032_000 || vest.NetSharesMicros != 2_636_000 || vest.WithheldSharesMicros != 1_396_000 {
		t.Fatalf("unexpected share totals: %#v", vest)
	}
	if vest.GrossValueCents != 128_657 {
		t.Fatalf("gross value = %d, want 128657", vest.GrossValueCents)
	}
}
