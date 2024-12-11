package toc

import (
	"encoding/csv"
	"fmt"
	"strings"
	"testing"
)

func TestLOL(t *testing.T) {
	// Input string
	line := `toc_signon        login.oscar.aol.com 5190 mike 0x230c0558310914310f english "pen\"guin haha" 'single quoted phrase'`

	// Use a CSV reader to parse the line
	reader := csv.NewReader(strings.NewReader(line))

	// Configure the CSV reader
	reader.Comma = ' '             // Use space as the field separator
	reader.LazyQuotes = true       // Allow unmatched quotes
	reader.TrimLeadingSpace = true // Allow unmatched quotes

	// Read the line
	records, err := reader.Read()
	if err != nil {
		fmt.Println("Error parsing line:", err)
		return
	}

	// Print the result
	fmt.Println(records)
}
