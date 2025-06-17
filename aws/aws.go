package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/xingbase/pigeon"
	"github.com/xingbase/pigeon/id"
)

const (
	DefaultRateTimeLimit = 500 * time.Second
	DefaultTemplateData  = "{}"
)

type Client struct {
	db *ses.Client

	id pigeon.ID

	region        string
	accessKeyID   string
	secretKey     string
	rateTimeLimit time.Duration
}

func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	c := &Client{
		db:            &ses.Client{},
		id:            id.NewTime(),
		region:        "us-west-1",
		rateTimeLimit: DefaultRateTimeLimit,
	}

	for i := range opts {
		if err := opts[i](c); err != nil {
			return nil, err
		}
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(c.region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(c.accessKeyID, c.secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("Error loading AWS config: %v\n", err)
	}

	c.db = ses.NewFromConfig(cfg)

	return c, nil
}

type Option func(*Client) error

func WithID(id pigeon.ID) Option {
	return func(c *Client) error {
		c.id = id
		return nil
	}
}

func WithRegion(s string) Option {
	return func(c *Client) error {
		c.region = s
		return nil
	}
}

func WithAccessKeyID(s string) Option {
	return func(c *Client) error {
		c.accessKeyID = s
		return nil
	}
}

func WithSecretKey(s string) Option {
	return func(c *Client) error {
		c.secretKey = s
		return nil
	}
}

func WithRateTimeLimit(d time.Duration) Option {
	return func(c *Client) error {
		if d < 0 {
			return nil
		}
		c.rateTimeLimit = d
		return nil
	}
}

func (c *Client) Email() pigeon.EmailSender {
	return &emailSender{c}
}
