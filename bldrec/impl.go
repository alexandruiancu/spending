package bldrec

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ProcessFiles(inDir, historyDir string) ([][]string, error) {
	var records [][]string

	// a) Open each text file in in_dir subdirectory
	files, err := os.ReadDir(inDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		filePath := filepath.Join(inDir, file.Name())

		// b) Read current file line by line and add to lines slice
		lines, err := readLines(filePath)
		if err != nil {
			return nil, err
		}

		// c) Move input file to history subdirectory
		historyPath := filepath.Join(historyDir, file.Name())
		if err := os.Rename(filePath, historyPath); err != nil {
			return nil, err
		}
		var aggregates [][]string
		for _, line := range lines {
			aggregate := regexp.MustCompile(`[ \t]{3,}`).Split(line, -1)
			if len(aggregate) > 0 {
				aggregates = append(aggregates, aggregate)
			}
		}
		var record []string
		for _, aggregate := range aggregates {
			if len(aggregate[0]) > 0 {
				if len(record) > 0 {
					records = append(records, record)
					record = nil
				}
				record = aggregate
			} else {
				for i, field := range aggregate {
					if len(field) == 0 {
						continue
					}
					record[i] += " " + field
				}
			}
		}
		if record != nil {
			records = append(records, record)
		}
	}

	return records, nil
}

func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
