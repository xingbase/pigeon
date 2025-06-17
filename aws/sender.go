package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/xingbase/pigeon"
)

var _ pigeon.EmailSender = (*emailSender)(nil)

type emailSender struct {
	*Client
}

func (e *emailSender) Send(ctx context.Context, msg pigeon.Message) error {
	tmpl := e.genTmpl()

	if err := e.mkTmpl(ctx, tmpl, msg.Subject, msg.Body); err != nil {
		return fmt.Errorf("Error creating template: %v\n", err)
	}
	defer func() {
		if err := e.rmTmpl(ctx, tmpl); err != nil {
			fmt.Printf("Error deleting template: %v\n", err)
		}
	}()

	dests := make([]types.BulkEmailDestination, 0, len(msg.To))
	for _, to := range msg.To {
		dests = append(dests, types.BulkEmailDestination{
			Destination: &types.Destination{
				ToAddresses: []string{to.Email},
			},
		})
	}

	req := &ses.SendBulkTemplatedEmailInput{
		Destinations:        dests,
		Source:              aws.String(string(msg.From)),
		Template:            aws.String(tmpl),
		DefaultTemplateData: aws.String(DefaultTemplateData),
	}

	resp, err := e.db.SendBulkTemplatedEmail(ctx, req)
	if err != nil {
		return fmt.Errorf("Error sending bulk email: %v\n", err)
	}

	for i, status := range resp.Status {
		if status.Status != types.BulkEmailStatusSuccess {
			fmt.Printf("Failed to send email to %s: MessageId=%s, Error=%s\n",
				msg.To[i].Email, aws.ToString(status.MessageId), aws.ToString(status.Error))
		}
	}

	return nil
}

func (e *emailSender) mkTmpl(ctx context.Context, key string, subject, body string) error {
	req := &ses.CreateTemplateInput{
		Template: &types.Template{
			TemplateName: aws.String(key),
			SubjectPart:  aws.String(subject),
			HtmlPart:     aws.String(body),
		},
	}

	if _, err := e.db.CreateTemplate(ctx, req); err != nil {
		return fmt.Errorf("failed to create template %s: %v", key, err)
	}

	return nil
}

func (e *emailSender) rmTmpl(ctx context.Context, key string) error {
	_, err := e.db.DeleteTemplate(ctx, &ses.DeleteTemplateInput{
		TemplateName: aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete template %s: %v", key, err)
	}

	return nil
}

func (e *emailSender) genTmpl() string {
	return fmt.Sprintf("TempEmailTemplate_%s", e.id.Generate())
}
