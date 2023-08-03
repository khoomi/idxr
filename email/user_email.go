package email

import (
	"fmt"
	"khoomi-api-io/khoomi_api/services"
	"log"
	"time"
)

type KhoomiEmailData struct {
	Email     string
	LoginName string
	Link      string
	LoginTime time.Time
	IP        string
}

func SendWelcomeEmail(email, loginName string) {
	mail := services.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     "no-reply@khoomi.com",
		SenderName: "Khoomi Online",
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>We are thrilled to welcome you to Khoomi, the new e-commerce platform for Nigerian creatives. Thank you for joining our community of talented artisans, small business owners, and passionate shoppers.</p><p>At Khoomi, our mission is to connect Nigerian creatives with a wider audience, helping you to showcase your products and services to a broader market. Our platform offers a user-friendly interface that allows you to easily list your products and reach customers across Nigeria and beyond.</p><p>As a new user, you can set up your seller account and start listing your products in just a few simple steps. Our team is always here to help you if you have any questions or concerns, so please don't hesitate to reach out.</p><p>We are excited to have you on board and can't wait to see the amazing products and services you have to offer. Thank you for choosing Khoomi!</p><p>Best,</p><p>The Khoomi Team</p></body></html>", loginName),
		Subject:    "Welcome to Khoomi - Connecting Nigerian Creatives to the World",
	}

	err := services.SendMail(mail)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("welcome email sent to %v", email)
	}

}

func SendVerifyEmailNotification(email, loginName, link string) {
	mail := services.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     "no-reply@khoomi.com",
		SenderName: "Khoomi Online",
		Body:       fmt.Sprintf(`<body style="font-family: Arial, sans-serif; font-size: 14px;"><p>Dear %v,</p><p>Thank you for creating an account with us!</p><p>Please click the following link to verify your email address:</p><p><a href="%v" style="color: #FF5810; text-decoration: none;">%v</a></p><p>The link is valid for 24 hours.</p><p>Best regards,</p><p>Your Application Team</p></body>`, loginName, link, link),
		Subject:    "Verify your email on Khoomi",
	}

	err := services.SendMail(mail)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("email verification notification email sent to %v", mail.To)
	}
}

func SendPasswordResetEmail(email, loginName, link string) {
	mail := services.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     "no-reply@khoomi.com",
		SenderName: "Khoomi Online",
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

	err := services.SendMail(mail)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("password reset email sent to %v", email)
	}
}

func SendPasswordResetSuccessfulEmail(email, loginName string) {
	mail := services.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     "no-reply@khoomi.com",
		SenderName: "Khoomi Online",
		Body:       fmt.Sprintf("<body style=\"font-family: Arial, sans-serif; font-size: 14px;><p>Dear %v,</p><p>Your password has been successfully reset.</p><p>If you did not request this password reset, please contact us immediately.</p><p>Best regards,</p><p>Your Application Team</p></body>", loginName),
		Subject:    "Password Reset Successful",
	}

	err := services.SendMail(mail)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("password reset successfully email sent to %v", email)
	}

}

func SendNewIpLoginNotification(email, loginName, IP string, loginTime time.Time) {
	mail := services.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     "no-reply@khoomi.com",
		SenderName: "Khoomi Online",
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>This is to inform you that a new IP address has been used to log in to your account at %v.</p><p>IP Address: %v</p> <p>If you did not log in from this location, please contact us or change your password immediately.</p><p>Best regards,</p><p>Your Application Team</p></body></html>", loginName, loginTime, IP),
		Subject:    "New IP Address Login Notification",
	}
	err := services.SendMail(mail)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("login from new IP addr email sent to %v", email)
	}
}
