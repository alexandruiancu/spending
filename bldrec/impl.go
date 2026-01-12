package bldrec

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"spending/common"
	"strconv"
	"strings"
	"sync"
	"time"

	capnp "capnproto.org/go/capnp/v3"
	zmq "github.com/pebbe/zmq4"
)

func Process() error {
	config := common.ReadConfig("../config.txt")

	inDir := config["in_dir"]
	historyDir := config["history_dir"]

	records, err := ProcessFiles(inDir, historyDir)
	if err != nil {
		return err
	}

	port := config["frontend_port"]
	for _, record := range records {
		socket, _ := zmq.NewSocket(zmq.REQ)
		defer socket.Close()
		socket.Connect(fmt.Sprintf("tcp://localhost:%s", port))
		// Serialize to a byte slice
		data, err := record.Message().Marshal()
		if err != nil {
			log.Fatalf("marshal: %v", err)
		}
		socket.SendBytes(data, 0)
		// Receive reply
		reply, _ := socket.Recv(0)
		fmt.Printf("Received reply: %s\n", reply)
	}

	return nil
}

func ProcessFiles(inDir, historyDir string) ([]Record, error) {
	var records []Record
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
						rec, err := createRecord(record)
						if err == nil {
							mu.Lock()
							records = append(records, rec)
							mu.Unlock()
						}
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
				rec, err := createRecord(record)
				if err == nil {
					mu.Lock()
					records = append(records, rec)
					mu.Unlock()
				}
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

func createRecord(fields []string) (Record, error) {
	// Create a new message arena
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		panic(err)
	}
	// Create a new Record
	rec, err := NewRootRecord(seg)
	if err != nil {
		return rec, err
	}

	// Set the fields from the string array
	if len(fields) > 0 {
		const layout = "02/01/2006"
		// Parse fields[0] as a date and convert to Unix timestamp (Int64)
		t, err := time.Parse(layout, fields[0])
		if err != nil {
			// Try alternative formats or fall back to current time
			t = time.Now()
		}
		rec.SetUDateTime(t.Unix())
	}
	if len(fields) > 1 {
		rec.SetSDescription(fields[1])
	}
	if len(fields) > 2 {
		amount := strings.ReplaceAll(fields[2], ",", ".")
		f, err := strconv.ParseFloat(amount, 32)
		if err != nil {
			f = 0.0
		}
		rec.SetFValue(float32(f))
	}
	if len(fields) > 3 {
		rec.SetSDontCare(fields[3])
	}

	return rec, nil
}
