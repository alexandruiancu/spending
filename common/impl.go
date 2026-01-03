package common

import (
	"bufio"
	"log"
	"os"
	"strings"
)

func ReadConfig(filepath string) map[string]string {
	config := make(map[string]string)
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return config
}
