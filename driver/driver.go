package main

import (
	"spending/bldrec"
	"strings"
)

func main() {
	// Example usage
	inDir := "../in_dir"
	historyDir := "../history"

	records, err := bldrec.ProcessFiles(inDir, historyDir)
	if err != nil {
		panic(err)
	}

	for _, record := range records {
		println(strings.Join(record, " | "))
	}
}
