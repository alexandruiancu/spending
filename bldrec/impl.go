package bldrec

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"spending/common"
	"strings"
	"sync"
)

func Process() error {
	config := common.ReadConfig("../config.txt")

	inDir := config["in_dir"]
	historyDir := config["history_dir"]

	records, err := ProcessFiles(inDir, historyDir)
	if err != nil {
		return err
	}

	for _, record := range records {
		println(strings.Join(record, " | "))
	}

	return nil
}

func ProcessFiles(inDir, historyDir string) ([][]string, error) {
	var records [][]string
	var wg sync.WaitGroup
	var mu sync.Mutex

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
		wg.Add(1)
		go func(filePath string) {
			defer wg.Done()

			lines, err := readLines(filePath)
			if err != nil {
				return
			}

			if err := os.Rename(filePath, filepath.Join(historyDir, filepath.Base(filePath))); err != nil {
				return
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
						mu.Lock()
						records = append(records, record)
						mu.Unlock()
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
				mu.Lock()
				records = append(records, record)
				mu.Unlock()
			}
		}(filePath)
	}
	wg.Wait()

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
