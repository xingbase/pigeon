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
