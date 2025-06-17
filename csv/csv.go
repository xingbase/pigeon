package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/xingbase/pigeon"
)

func Load(path string, skipRows int, out chan<- pigeon.To) error {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)

	line := skipRows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV record at line %d: %v", line, err)
		}
		if len(record) < 1 {
			return fmt.Errorf("invalid record at line %d: expected 1 column, got %d", line, len(record))
		}

		out <- pigeon.To{Email: record[0]}
		line++
	}
	return nil
}

func Write(filename string, record []string, isHeader bool) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", filename, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if isHeader {
		// Write header only if file is new
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat %s: %v", filename, err)
		}
		if fileInfo.Size() == 0 {
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write header to %s: %v", filename, err)
			}
			return nil
		}
	} else {
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write to %s: %v", filename, err)
		}
	}

	return nil
}
