package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"life/internal/models"
	"life/internal/payroll"
	"life/internal/storage"
)

func main() {
	databasePath := flag.String("db", "life.db", "path to the Life SQLite database")
	yes := flag.Bool("yes", false, "import without an interactive confirmation")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: import-payroll [--db life.db] [--yes] PAY_STATEMENT.pdf")
		os.Exit(2)
	}

	statement, err := payroll.ParsePDF(flag.Arg(0))
	if err != nil {
		fatal(err)
	}

	printPreview(statement)
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
	if _, err := db.AddPayrollStatement(statement); err != nil {
		if errors.Is(err, storage.ErrPayrollAlreadyImported) {
			fatal(fmt.Errorf("this pay statement has already been imported"))
		}
		fatal(err)
	}

	fmt.Println("Payroll statement imported successfully.")
}

func printPreview(s models.PayrollStatement) {
	fmt.Println("PAYROLL IMPORT PREVIEW")
	fmt.Printf("Pay period:  %s - %s\n", s.PeriodStart.Format("Jan 2, 2006"), s.PeriodEnd.Format("Jan 2, 2006"))
	fmt.Printf("Pay date:    %s\n", s.PayDate.Format("Jan 2, 2006"))
	fmt.Printf("Gross pay:   %s\n", formatCents(s.GrossCents))
	fmt.Printf("Taxable pay: %s\n", formatCents(s.TaxableCents))
	fmt.Printf("Taxes:       %s\n", formatCents(s.TaxesCents))
	fmt.Printf("Deductions:  %s\n", formatCents(s.DeductionsCents))
	fmt.Printf("Net pay:     %s\n", formatCents(s.NetCents))
	fmt.Println("\nNames, addresses, employee IDs, SSNs, and bank details will not be saved.")
}

func confirmImport() bool {
	fmt.Print("Import these values? [y/N]: ")
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(answer), "y")
}

func formatCents(cents int64) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(1)
}
