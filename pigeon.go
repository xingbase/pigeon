package pigeon

import (
	"context"
)

type From string

type To struct {
	Name  *string
	Email string
}

type Message struct {
	From    From
	To      []To
	Subject string
	Body    string
}

type EmailSender interface {
	Send(context.Context, Message) error
}

// ID creates uniq ID string
type ID interface {
	// Generate creates a unique ID string
	Generate() string
}
