package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xingbase/pigeon"
	"github.com/xingbase/pigeon/aws"
	"github.com/xingbase/pigeon/csv"
)

func init() {
	parser.AddCommand("aws",
		"with aws ses email",
		"with aws ses send to email",
		&sendCommand)
}

type SendCommand struct {
	CsvFile     string `short:"f" long:"file" description:"path to csv file" env:"PIGEON_CSV_FILE"`
	CsvSkipRows int    `long:"skip-rows" description:"skip rows" default:"1" env:"PIGEON_CSV_SKIP_ROWS"`

	Sender    string `short:"s" long:"sender" description:"sender email address" env:"PIGEON_SENDER"`
	Subject   string `short:"t" long:"subject" description:"email subject" env:"PIGEON_SUBJECT"`
	Body      string `short:"b" long:"body" description:"email body" env:"PIGEON_BODY"`
	BatchSize int    `short:"n" long:"batch-size" description:"batch size" default:"50" env:"PIGEON_BATCH_SIZE"`

	AwsRegion        string        `long:"region" description:"AWS region" env:"AWS_REGION"`
	AwsAccessKeyID   string        `long:"access-key-id" description:"AWS access key ID" env:"AWS_ACCESS_KEY_ID"`
	AwsSecretKey     string        `long:"secret-key" description:"AWS secret key" env:"AWS_SECRET_KEY"`
	AwsRateTimeLimit time.Duration `long:"rate-time-limit" description:"AWS rate time limit" default:"500s" env:"AWS_RATE_TIME_LIMIT"`
}

var sendCommand SendCommand

func (s *SendCommand) Execute(args []string) error {
	client, err := aws.NewClient(context.Background(),
		aws.WithRegion(s.AwsRegion),
		aws.WithAccessKeyID(s.AwsAccessKeyID),
		aws.WithSecretKey(s.AwsSecretKey),
		aws.WithRateTimeLimit(s.AwsRateTimeLimit),
	)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	toChan := make(chan pigeon.To, s.BatchSize)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := csv.Load(s.CsvFile, 1, toChan); err != nil {
			fmt.Printf("Error loading CSV file: %v\n", err)
			cancel()
		}
		close(toChan)
	}()

	batch := make([]pigeon.To, 0, s.BatchSize)
	total := 0

	for {
		select {
		case to, ok := <-toChan:
			if !ok {
				// Channel closed, process any remaining recipients
				if len(batch) > 0 {
					if err := client.Email().Send(ctx, pigeon.Message{
						From:    pigeon.From(s.Sender),
						To:      batch,
						Subject: s.Subject,
						Body:    s.Body,
					}); err != nil {
						fmt.Printf("Error sending email: %v\n", err)
					}
					total += len(batch)
				}
				goto Done
			}
			batch = append(batch, to)
			if len(batch) == s.BatchSize {
				if err := client.Email().Send(ctx, pigeon.Message{
					From:    pigeon.From(s.Sender),
					To:      batch,
					Subject: s.Subject,
					Body:    s.Body,
				}); err != nil {
					fmt.Printf("Error sending email: %v\n", err)
				}
			}
		case <-ctx.Done():
			goto Done
		}
	}

Done:
	wg.Wait()
	fmt.Printf("Email sending completed, processed %d recipients\n", total)

	return nil
}
