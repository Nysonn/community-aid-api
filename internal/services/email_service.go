package services

import (
	"fmt"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client *resend.Client
}

func NewEmailService(client *resend.Client) *EmailService {
	return &EmailService{client: client}
}

func (s *EmailService) send(to, subject, body string) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    "CommunityAid <noreply@communityaid.com>",
		To:      []string{to},
		Subject: subject,
		Text:    body,
	})
	return err
}

func (s *EmailService) SendRequestApprovedEmail(to, requestTitle string) error {
	subject := "Your emergency request has been approved"
	body := fmt.Sprintf(
		"Good news! Your emergency request titled \"%s\" has been reviewed and approved. "+
			"It is now visible to the community and members can offer assistance.\n\n"+
			"Thank you for using CommunityAid.",
		requestTitle,
	)
	return s.send(to, subject, body)
}

func (s *EmailService) SendRequestRejectedEmail(to, requestTitle string) error {
	subject := "Your emergency request was not approved"
	body := fmt.Sprintf(
		"We regret to inform you that your emergency request titled \"%s\" was reviewed "+
			"and could not be approved at this time.\n\n"+
			"If you believe this is an error, please contact our support team.\n\n"+
			"Thank you for using CommunityAid.",
		requestTitle,
	)
	return s.send(to, subject, body)
}

func (s *EmailService) SendOfferNotificationEmail(to, requestTitle string) error {
	subject := "Someone has offered help for your request"
	body := fmt.Sprintf(
		"Great news! Someone has offered assistance for your emergency request titled \"%s\".\n\n"+
			"Please log in to CommunityAid to view the offer details and get in touch with the responder.\n\n"+
			"Thank you for using CommunityAid.",
		requestTitle,
	)
	return s.send(to, subject, body)
}
