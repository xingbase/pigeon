package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pcsv "github.com/xingbase/pigeon/csv"
	pmail "github.com/xingbase/pigeon/mail"
)

func init() {
	parser.AddCommand("domain",
		"Check Domain Validation",
		"Check Domain Validation",
		&domainCommand)
}

type DomainCommand struct {
	CsvFile string `short:"f" long:"file" description:"CSV file" required:"true"`
}

var domainCommand DomainCommand

func (s *DomainCommand) Execute(args []string) error {
	// Initialize error log file
	logFile, err := os.Create("errors.log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	// Open input CSV file
	file, err := os.Open(s.CsvFile)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.Comma = '\t'
	domainCh := make(chan string, 100)
	workerCount := 10 // Adjust based on system and DNS server capacity

	// Read CSV and send domains to channel
	var totalCount int64
	go func() {
		defer close(domainCh)
		for {
			record, err := reader.Read()
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				break
			}
			if len(record) == 0 || record[0] == "" {
				continue
			}
			// Skip header
			if strings.Contains(strings.ToLower(record[0]), "domain") {
				continue
			}
			atomic.AddInt64(&totalCount, 1) // Increment domain count
			domainCh <- record[0]
		}
	}()

	// Wait briefly to ensure some domains are counted
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt64(&totalCount) == 0 {
		logger.Fatal("No domains found in CSV file")
	}

	// Process domains
	validCount, invalidCount := processDomains(domainCh, workerCount, atomic.LoadInt64(&totalCount))

	fmt.Printf("Processed %d domains: %d valid, %d invalid\n", atomic.LoadInt64(&totalCount), validCount, invalidCount)
	return nil
}

type DomainResult struct {
	Domain string
	Valid  bool
	Error  string
}

// processDomains validates domains concurrently and streams results to CSV files.
func processDomains(domains <-chan string, workerCount int, totalCount int64) (int64, int64) {
	var validCount, invalidCount int64
	var wg sync.WaitGroup
	var mu sync.Mutex
	var processed int64
	resultCh := make(chan DomainResult, workerCount)

	// Initialize CSV files with headers
	if err := pcsv.Write("domains_valid.csv", []string{"Domain"}, true); err != nil {
		log.Fatalf("Error initializing domains_valid.csv: %v", err)
	}
	if err := pcsv.Write("domains_invalid.csv", []string{"Domain", "Error"}, true); err != nil {
		log.Fatalf("Error initializing domains_invalid.csv: %v", err)
	}

	// Start worker pool
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for domain := range domains {
				valid, err := pmail.IsValidDomain(domain)
				result := DomainResult{
					Domain: domain,
					Valid:  valid,
				}
				if err != nil {
					result.Error = err.Error()
				}
				resultCh <- result
			}
		}()
	}

	// Stream results to CSV in a separate goroutine
	go func() {
		for result := range resultCh {
			mu.Lock()
			if result.Valid {
				if err := pcsv.Write("domains_valid.csv", []string{result.Domain}, false); err != nil {
					log.Printf("Error writing to domains_valid.csv: %v", err)
				}
				atomic.AddInt64(&validCount, 1)
			} else {
				if err := pcsv.Write("domains_invalid.csv", []string{result.Domain, result.Error}, false); err != nil {
					log.Printf("Error writing to domains_invalid.csv: %v", err)
				}
				atomic.AddInt64(&invalidCount, 1)
			}

			// Update progress
			count := atomic.AddInt64(&processed, 1)
			if count%100 == 0 || count == totalCount {
				fmt.Printf("Processed %d/%d domains (%.2f%%)\n", count, totalCount, float64(count)/float64(totalCount)*100)
			}
			mu.Unlock()
		}
	}()

	// Wait for all workers to finish and close result channel
	wg.Wait()
	close(resultCh)

	return validCount, invalidCount
}
