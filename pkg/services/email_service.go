package services

import (
	"fmt"
	"khoomi-api-io/api/pkg/util"
	"log"
	"time"
)

type emailService struct {
	sender     string
	senderName string
}

// NewEmailService creates a new instance of EmailService
func NewEmailService() EmailService {
	return &emailService{
		sender:     "no-reply@khoomi.com",
		senderName: "Khoomi Online",
	}
}

// SendWelcomeEmail sends a welcome message to new users after they register an account with Khoomi.
func (e *emailService) SendWelcomeEmail(email, loginName string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>We are thrilled to welcome you to Khoomi, the new e-commerce platform for Nigerian creatives. Thank you for joining our community of talented artisans, small business owners, and passionate shoppers.</p><p>At Khoomi, our mission is to connect Nigerian creatives with a wider audience, helping you to showcase your products and services to a broader market. Our platform offers a user-friendly interface that allows you to easily list your products and reach customers across Nigeria and beyond.</p><p>As a new user, you can set up your seller account and start listing your products in just a few simple steps. Our team is always here to help you if you have any questions or concerns, so please don't hesitate to reach out.</p><p>We are excited to have you on board and can't wait to see the amazing products and services you have to offer. Thank you for choosing Khoomi!</p><p>Best,</p><p>The Khoomi Team</p></body></html>", loginName),
		Subject:    "Welcome to Khoomi - Connecting Nigerian Creatives to the World",
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send welcome email:", err)
		return err
	}
	log.Printf("Welcome email sent to %v", email)
	return nil
}

// SendVerifyEmailNotification constructs and sends an email to a user with a link to verify their email address.
func (e *emailService) SendVerifyEmailNotification(email, loginName, link string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf(`<body style="font-family: Arial, sans-serif; font-size: 14px;"><p>Dear %v,</p><p>Thank you for creating an account with us!</p><p>Please click the following link to verify your email address:</p><p><a href="%v" style="color: #FF5810; text-decoration: none;">%v</a></p><p>The link is valid for 24 hours.</p><p>Best regards,</p><p>Your Application Team</p></body>`, loginName, link, link),
		Subject:    "Verify your email on Khoomi",
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send email verification notification:", err)
		return err
	}
	log.Printf("Email verification notification sent to %v", email)
	return nil
}

// SendEmailVerificationSuccessNotification sends a notification email to a user
// confirming that their email address has been successfully verified.
func (e *emailService) SendEmailVerificationSuccessNotification(email, loginName string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf(`<body style="font-family: Arial, sans-serif; font-size: 14px;"><p>Dear %v,</p><p>Your email has been successfully verified!</p><p>Thank you for verifying your email address. You can now enjoy full access to our util.</p><p>Best regards,</p><p>Your Khoomi Team</p></body>`, loginName),
		Subject:    "Email Verification Successful",
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send verification success email:", err)
		return err
	}
	log.Printf("Verification success email sent to %v", email)
	return nil
}

// SendPasswordResetEmail constructs and sends a password reset email to a user.
func (e *emailService) SendPasswordResetEmail(email, loginName, link string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<body style="font-family: Arial, sans-serif; font-size: 14px;">
    <p>Dear %v,</p>
    <p>We received a request to reset the password for your Khoomi account. To reset your password, please click the button below:</p>
    <div style="text-align: center;">
        <a href="%v" style="background-color: #FF5810; color: #ffffff; border-radius: 30px; display: inline-block; font-size: 16px; font-weight: bold; padding: 10px 16px; text-align: center; text-decoration: none;">Reset Password</a>
    </div>
    <p>If you did not request a password reset, please ignore this message and your password will remain unchanged.</p>
    <p>Thank you,</p>
    <p>The Khoomi Team</p>
</body>`, loginName, link),
		Subject: "Khoomi Password Reset Request",
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send password reset email:", err)
		return err
	}
	log.Printf("Password reset email sent to %v", email)
	return nil
}

// SendPasswordResetSuccessfulEmail sends a notification email to a user to confirm that their password has been successfully reset.
func (e *emailService) SendPasswordResetSuccessfulEmail(email, loginName string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf(`<body style="font-family: Arial, sans-serif; font-size: 14px;"><p>Dear %v,</p><p>Your password has been successfully reset.</p><p>If you did not request this password reset, please contact us immediately.</p><p>Best regards,</p><p>Your Application Team</p></body>`, loginName),
		Subject:    "Password Reset Successful",
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send password reset successful email:", err)
		return err
	}
	log.Printf("Password reset successful email sent to %v", email)
	return nil
}

// SendNewIpLoginNotification sends an alert email to a user when a login attempt is made from a new IP address.
func (e *emailService) SendNewIpLoginNotification(email, loginName, IP string, loginTime time.Time) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>This is to inform you that a new IP address has been used to log in to your account at %v.</p><p>IP Address: %v</p> <p>If you did not log in from this location, please contact us or change your password immediately.</p><p>Best regards,</p><p>Your Application Team</p></body></html>", loginName, loginTime, IP),
		Subject:    "New IP Address Login Notification",
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send new IP login notification:", err)
		return err
	}
	log.Printf("New IP login notification sent to %v", email)
	return nil
}

// SendNewShopEmail sends a notification email to a user when their shop is successfully created on Khoomi.
func (e *emailService) SendNewShopEmail(email, sellerName, shopName string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     sellerName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>Congratulations on creating a new shop on Khoomi!</p><p>Your shop named <span class=\"highlight\">'%v'</span> has been successfully created and is now live on our platform. This means that your products and services are now accessible to a wide audience of passionate shoppers.</p><p>We wish you the best of luck with your new shop. If you have any questions or need assistance, please feel free to reach out to our support team.</p><p>Thank you for choosing Khoomi as your e-commerce partner. We look forward to seeing your business thrive on our platform.</p><p>Best,</p><p>The Khoomi Team</p></body></html>", sellerName, shopName),
		Subject:    fmt.Sprintf("Congratulations on Your New Shop - %v", shopName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send new shop email:", err)
		return err
	}
	log.Printf("New shop email sent to %v", email)
	return nil
}

// SendNewListingEmail sends a notification email to a user when a new listing is successfully posted in their shop on Khoomi.
func (e *emailService) SendNewListingEmail(email, sellerName, listingTitle string) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     sellerName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>Congratulations on creating a new listing on Khoomi!</p><p>Your listing titled <span class=\"highlight\">'%v'</span> has been successfully created and is now live on our platform. This means that your products and services are now accessible to a wide audience of passionate shoppers.</p><p>We wish you the best of luck with your new listing. If you have any questions or need assistance, please feel free to reach out to our support team.</p><p>Thank you for choosing Khoomi as your e-commerce partner. We look forward to seeing your business thrive on our platform.</p><p>Best,</p><p>The Khoomi Team</p></body></html>", sellerName, listingTitle),
		Subject:    fmt.Sprintf("Congratulations on Your New Listing - %v", listingTitle),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send new listing email:", err)
		return err
	}
	log.Printf("New listing email sent to %v", email)
	return nil
}