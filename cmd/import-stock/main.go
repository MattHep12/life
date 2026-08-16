package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"life/internal/stock"
	"life/internal/storage"
)

func main() {
	databasePath := flag.String("db", "life.db", "path to the Life SQLite database")
	yes := flag.Bool("yes", false, "import without an interactive confirmation")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: import-stock [--db life.db] [--yes] SCHWAB.tsv")
		os.Exit(2)
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	vests, err := stock.ParseTSV(file)
	_ = file.Close()
	if err != nil {
		fatal(err)
	}
	if len(vests) == 0 {
		fatal(fmt.Errorf("no restricted-stock rows found"))
	}

	var grossValue, grossShares, netShares int64
	for _, vest := range vests {
		grossValue += vest.GrossValueCents
		grossShares += vest.GrossSharesMicros
		netShares += vest.NetSharesMicros
	}
	fmt.Println("STOCK VEST IMPORT PREVIEW")
	fmt.Printf("Vest events:       %d\n", len(vests))
	fmt.Printf("Gross shares:      %s\n", stock.FormatShares(grossShares))
	fmt.Printf("Net shares:        %s\n", stock.FormatShares(netShares))
	fmt.Printf("Withheld shares:   %s\n", stock.FormatShares(grossShares-netShares))
	fmt.Printf("Gross vest value:  $%d.%02d\n", grossValue/100, grossValue%100)
	fmt.Println("Award IDs, award dates, market-value snapshots, and dividend reinvestments will not be saved.")

	if !*yes && !confirmImport() {
		fmt.Println("Import canceled; nothing was saved.")
		return
	}

	db, err := storage.Open(*databasePath)
	if err != nil {
		fatal(err)
	}
	defer db.Close()
	if err := db.CreateTables(); err != nil {
		fatal(err)
	}

	imported, skipped := 0, 0
	for _, vest := range vests {
		if _, err := db.AddStockVest(vest); err != nil {
			if errors.Is(err, storage.ErrStockVestAlreadyImported) {
				skipped++
				continue
			}
			fatal(err)
		}
		imported++
	}
	fmt.Printf("Imported %d stock vest events; skipped %d duplicates.\n", imported, skipped)
}

func confirmImport() bool {
	fmt.Print("Import these values? [y/N]: ")
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(answer), "y")
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
