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
		From:    "CommunityAid <no-reply@communityaid.me>",
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

// SendDonationReceivedEmail notifies the request owner that a donation was collected by admin
// and is being processed for disbursement to their account.
func (s *EmailService) SendDonationReceivedEmail(to, requestTitle, donorName string, amount float64) error {
	subject := "A donation has been received for your request"
	body := fmt.Sprintf(
		"Good news! A donation of UGX %.0f from %s has been received by the CommunityAid admin "+
			"for your request \"%s\".\n\n"+
			"The admin will disburse these funds to your registered payment account shortly.\n\n"+
			"Thank you for using CommunityAid.",
		amount, donorName, requestTitle,
	)
	return s.send(to, subject, body)
}

// SendFundsDisbursedToRecipientEmail notifies the request owner that funds have been sent to their account.
func (s *EmailService) SendFundsDisbursedToRecipientEmail(to, requestTitle string, amount float64) error {
	subject := "Funds have been sent to your account"
	body := fmt.Sprintf(
		"The CommunityAid admin has successfully disbursed UGX %.0f to your registered payment account "+
			"for your request \"%s\".\n\n"+
			"Please check your account to confirm receipt.\n\n"+
			"Thank you for using CommunityAid.",
		amount, requestTitle,
	)
	return s.send(to, subject, body)
}

// SendDonorFundsDeliveredEmail notifies the donor that their funds have been forwarded to the recipient.
func (s *EmailService) SendDonorFundsDeliveredEmail(to, requestTitle, recipientName string, amount float64) error {
	subject := "Your donation has reached the recipient"
	body := fmt.Sprintf(
		"Great news! Your donation of UGX %.0f for the request \"%s\" has been successfully "+
			"forwarded to %s by the CommunityAid admin.\n\n"+
			"Thank you for your generosity and for using CommunityAid.",
		amount, requestTitle, recipientName,
	)
	return s.send(to, subject, body)
}
