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

// createEmailTemplate creates a simple, clean email template
func (e *emailService) createEmailTemplate(title, content string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            font-size: 16px;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            padding: 20px 0;
            border-bottom: 1px solid #eee;
            margin-bottom: 30px;
        }
        
        .header h1 {
            color: #FF5810;
            margin: 0;
            font-size: 24px;
        }
        
        .content p {
            margin-bottom: 16px;
        }
        
        .highlight {
            color: #FF5810;
            font-weight: bold;
        }
        
        .button {
            display: inline-block;
            background-color: #FF5810;
            color: white;
            text-decoration: none;
            padding: 12px 24px;
            border-radius: 4px;
            margin: 20px 0;
        }
        
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            font-size: 14px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Khoomi</h1>
    </div>
    
    <div class="content">
        %s
    </div>
    
    <div class="footer">
        <p>Best regards,<br>The Khoomi Team</p>
        <p><a href="https://khoomi.com">khoomi.com</a></p>
    </div>
</body>
</html>`, title, content)
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Welcome to Khoomi! We are excited to have you join our community of Nigerian creatives.</p>
		
		<p>You can now set up your seller account and start listing your products. Our platform makes it easy to connect with customers across Nigeria and beyond.</p>
		
		<p>If you have any questions, our support team is here to help.</p>
		
		<p>Thank you for choosing Khoomi!</p>
	`, loginName)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Welcome to Khoomi", content),
		Subject:    "Welcome to Khoomi",
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Please verify your email address to activate your Khoomi account.</p>
		
		<p><a href="%s" class="button">Verify Email Address</a></p>
		
		<p>This link expires in 24 hours.</p>
		
		<p>If you didn't create this account, please ignore this email.</p>
	`, loginName, link)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Verify Your Email", content),
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Your email has been successfully verified!</p>
		
		<p>You can now enjoy full access to your Khoomi account.</p>
		
		<p>Thank you for verifying your email address.</p>
	`, loginName)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Email Verification Successful", content),
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>We received a request to reset your Khoomi password.</p>
		
		<p><a href="%s" class="button">Reset Password</a></p>
		
		<p>This link is valid for 1 hour.</p>
		
		<p>If you didn't request this password reset, please ignore this email.</p>
	`, loginName, link)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Password Reset Request", content),
		Subject:    "Reset your Khoomi password",
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Your password has been successfully reset.</p>
		
		<p>If you did not request this password reset, please contact us immediately.</p>
	`, loginName)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Password Reset Successful", content),
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>A new IP address has been used to log in to your account.</p>
		
		<p><strong>Login Time:</strong> %s<br>
		<strong>IP Address:</strong> %s</p>
		
		<p>If you did not log in from this location, please contact us or change your password immediately.</p>
	`, loginName, loginTime.Format("January 2, 2006 at 3:04 PM"), IP)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     loginName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("New IP Login Alert", content),
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Congratulations! Your shop has been successfully created on Khoomi.</p>
		
		<p><strong>Shop Name:</strong> %s</p>
		
		<p>Your shop is now live and ready to reach customers across Nigeria and beyond.</p>
		
		<p>If you need help getting started, our support team is here to assist you.</p>
	`, sellerName, shopName)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     sellerName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Shop Created Successfully", content),
		Subject:    fmt.Sprintf("Your shop %s is now live!", shopName),
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
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Your new listing has been successfully created and is now live on Khoomi.</p>
		
		<p><strong>Listing:</strong> %s</p>
		
		<p>Customers can now view and purchase this item from your shop.</p>
	`, sellerName, listingTitle)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     sellerName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("New Listing Created", content),
		Subject:    fmt.Sprintf("New listing created: %s", listingTitle),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send new listing email:", err)
		return err
	}
	log.Printf("New listing email sent to %v", email)
	return nil
}

// SendShopNewOrderNotification sends an email notification when a new order is received
func (e *emailService) SendShopNewOrderNotification(email, shopName, orderID, customerName string, orderTotal float64) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>You have received a new order on Khoomi.</p>
		
		<p><strong>Order ID:</strong> #%s<br>
		<strong>Customer:</strong> %s<br>
		<strong>Total:</strong> ₦%.2f</p>
		
		<p>Please log in to your dashboard to process this order.</p>
	`, shopName, orderID, customerName, orderTotal)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("New Order Received", content),
		Subject:    fmt.Sprintf("New Order #%s", orderID),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send new order notification:", err)
		return err
	}
	log.Printf("New order notification sent to %v", email)
	return nil
}

// SendShopPaymentConfirmedNotification sends an email when payment is confirmed
func (e *emailService) SendShopPaymentConfirmedNotification(email, shopName, orderID string, amount float64) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Payment has been confirmed for your order.</p>
		
		<p><strong>Order ID:</strong> #%s<br>
		<strong>Amount:</strong> ₦%.2f</p>
		
		<p>You can now proceed with fulfilling this order.</p>
	`, shopName, orderID, amount)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Payment Confirmed", content),
		Subject:    fmt.Sprintf("Payment Confirmed - Order #%s", orderID),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send payment confirmed notification:", err)
		return err
	}
	log.Printf("Payment confirmed notification sent to %v", email)
	return nil
}

// SendShopPaymentFailedNotification sends an email when payment fails
func (e *emailService) SendShopPaymentFailedNotification(email, shopName, orderID, reason string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Payment failed for an order in your shop.</p>
		
		<p><strong>Order ID:</strong> #%s<br>
		<strong>Reason:</strong> %s</p>
		
		<p>Please check your seller dashboard for more details.</p>
	`, shopName, orderID, reason)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Payment Failed", content),
		Subject:    fmt.Sprintf("Payment Failed - Order #%s", orderID),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send payment failed notification:", err)
		return err
	}
	log.Printf("Payment failed notification sent to %v", email)
	return nil
}

// SendShopOrderCancelledNotification sends an email when an order is cancelled
func (e *emailService) SendShopOrderCancelledNotification(email, shopName, orderID, customerName string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>An order has been cancelled in your shop.</p>
		
		<p><strong>Order ID:</strong> #%s<br>
		<strong>Customer:</strong> %s</p>
		
		<p>Please check your seller dashboard for more details about this cancellation.</p>
	`, shopName, orderID, customerName)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Order Cancelled", content),
		Subject:    fmt.Sprintf("Order Cancelled - Order #%s", orderID),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send order cancelled notification:", err)
		return err
	}
	log.Printf("Order cancelled notification sent to %v", email)
	return nil
}

// SendShopLowStockNotification sends an email when product stock is low
func (e *emailService) SendShopLowStockNotification(email, shopName, productName string, currentStock int, threshold int) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Low stock alert! One of your products is running low on inventory.</p>
		
		<p><strong>Product:</strong> %s<br>
		<strong>Current Stock:</strong> %d units<br>
		<strong>Threshold:</strong> %d units</p>
		
		<p>Consider restocking this product to avoid going out of stock.</p>
	`, shopName, productName, currentStock, threshold)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Low Stock Alert", content),
		Subject:    fmt.Sprintf("Low Stock Alert - %s", productName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send low stock notification:", err)
		return err
	}
	log.Printf("Low stock notification sent to %v", email)
	return nil
}

// SendShopOutOfStockNotification sends an email when a product is out of stock
func (e *emailService) SendShopOutOfStockNotification(email, shopName, productName string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Out of stock alert! One of your products is now completely out of stock.</p>
		
		<p><strong>Product:</strong> %s<br>
		<strong>Current Stock:</strong> 0 units</p>
		
		<p>This product is no longer available for purchase. Please restock as soon as possible.</p>
	`, shopName, productName)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Out of Stock Alert", content),
		Subject:    fmt.Sprintf("Out of Stock Alert - %s", productName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send out of stock notification:", err)
		return err
	}
	log.Printf("Out of stock notification sent to %v", email)
	return nil
}

// SendShopInventoryRestockedNotification sends an email when inventory is restocked
func (e *emailService) SendShopInventoryRestockedNotification(email, shopName, productName string, newStock int) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Inventory restocked! Your product inventory has been updated.</p>
		
		<p><strong>Product:</strong> %s<br>
		<strong>New Stock Level:</strong> %d units</p>
		
		<p>Your product is now available for purchase again.</p>
	`, shopName, productName, newStock)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Inventory Restocked", content),
		Subject:    fmt.Sprintf("Inventory Restocked - %s", productName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send inventory restocked notification:", err)
		return err
	}
	log.Printf("Inventory restocked notification sent to %v", email)
	return nil
}

// SendShopNewReviewNotification sends an email when a new review is posted
func (e *emailService) SendShopNewReviewNotification(email, shopName, productName, reviewerName string, rating int) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>You have received a new review on one of your products.</p>
		
		<p><strong>Product:</strong> %s<br>
		<strong>Reviewer:</strong> %s<br>
		<strong>Rating:</strong> %d/5 stars</p>
		
		<p>Check your seller dashboard to read the full review and respond if needed.</p>
	`, shopName, productName, reviewerName, rating)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("New Review Received", content),
		Subject:    fmt.Sprintf("New Review Received - %s", productName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send new review notification:", err)
		return err
	}
	log.Printf("New review notification sent to %v", email)
	return nil
}

// SendShopCustomerMessageNotification sends an email when a customer sends a message
func (e *emailService) SendShopCustomerMessageNotification(email, shopName, customerName, subject string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>You have received a new message from a customer.</p>
		
		<p><strong>From:</strong> %s<br>
		<strong>Subject:</strong> %s</p>
		
		<p>Please log in to your seller dashboard to read and respond to this message.</p>
	`, shopName, customerName, subject)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("New Customer Message", content),
		Subject:    fmt.Sprintf("New Customer Message - %s", subject),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send customer message notification:", err)
		return err
	}
	log.Printf("Customer message notification sent to %v", email)
	return nil
}

// SendShopReturnRequestNotification sends an email when a return request is made
func (e *emailService) SendShopReturnRequestNotification(email, shopName, orderID, customerName, reason string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>A customer has requested to return an order.</p>
		
		<p><strong>Order ID:</strong> #%s<br>
		<strong>Customer:</strong> %s<br>
		<strong>Reason:</strong> %s</p>
		
		<p>Please review this return request in your seller dashboard and take appropriate action.</p>
	`, shopName, orderID, customerName, reason)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Return Request", content),
		Subject:    fmt.Sprintf("Return Request - Order #%s", orderID),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send return request notification:", err)
		return err
	}
	log.Printf("Return request notification sent to %v", email)
	return nil
}

// SendShopSalesSummaryNotification sends periodic sales summary emails
func (e *emailService) SendShopSalesSummaryNotification(email, shopName string, period string, totalSales float64, orderCount int) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Here's your sales summary for the %s:</p>
		
		<p><strong>Total Revenue:</strong> ₦%.2f<br>
		<strong>Total Orders:</strong> %d orders</p>
		
		<p>Keep up the great work! Check your seller dashboard for detailed analytics.</p>
	`, shopName, period, totalSales, orderCount)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Sales Summary", content),
		Subject:    fmt.Sprintf("Sales Summary - %s", period),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send sales summary notification:", err)
		return err
	}
	log.Printf("Sales summary notification sent to %v", email)
	return nil
}

// SendShopRevenueMilestoneNotification sends an email when a revenue milestone is reached
func (e *emailService) SendShopRevenueMilestoneNotification(email, shopName string, milestone float64, period string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Congratulations! You've reached a new revenue milestone!</p>
		
		<p><strong>Milestone Amount:</strong> ₦%.2f<br>
		<strong>Achievement Period:</strong> %s</p>
		
		<p>This is a fantastic achievement! Keep growing your business on Khoomi.</p>
	`, shopName, milestone, period)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Revenue Milestone Reached", content),
		Subject:    fmt.Sprintf("Revenue Milestone Reached - ₦%.2f", milestone),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send revenue milestone notification:", err)
		return err
	}
	log.Printf("Revenue milestone notification sent to %v", email)
	return nil
}

// SendShopPopularProductNotification sends an email about trending products
func (e *emailService) SendShopPopularProductNotification(email, shopName, productName string, salesCount int, period string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Trending product alert! Your product is performing exceptionally well.</p>
		
		<p><strong>Product:</strong> %s<br>
		<strong>Sales in %s:</strong> %d units</p>
		
		<p>Consider promoting this product more or creating similar items to capitalize on this trend.</p>
	`, shopName, productName, period, salesCount)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Trending Product", content),
		Subject:    fmt.Sprintf("Trending Product - %s", productName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send popular product notification:", err)
		return err
	}
	log.Printf("Popular product notification sent to %v", email)
	return nil
}

// SendShopAccountVerificationNotification sends account verification status updates
func (e *emailService) SendShopAccountVerificationNotification(email, shopName, status string) error {
	var statusMessage string
	switch status {
	case "approved":
		statusMessage = "Your shop has been successfully verified!"
	case "pending":
		statusMessage = "Your shop verification is under review."
	case "rejected":
		statusMessage = "Your shop verification was not approved."
	default:
		statusMessage = "Your shop verification status has been updated."
	}

	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Your account verification status has been updated.</p>
		
		<p><strong>Status:</strong> %s<br>
		<strong>Message:</strong> %s</p>
		
		<p>Check your seller dashboard for more details and next steps.</p>
	`, shopName, status, statusMessage)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Account Verification Update", content),
		Subject:    fmt.Sprintf("Account Verification Update - %s", shopName),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send account verification notification:", err)
		return err
	}
	log.Printf("Account verification notification sent to %v", email)
	return nil
}

// SendShopPolicyUpdateNotification sends notifications about policy changes
func (e *emailService) SendShopPolicyUpdateNotification(email, shopName, policyType, summary string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>We have updated our policy.</p>
		
		<p><strong>Policy Type:</strong> %s<br>
		<strong>Summary of Changes:</strong> %s</p>
		
		<p>Please review the updated policy in your seller dashboard to ensure compliance.</p>
	`, shopName, policyType, summary)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Policy Update", content),
		Subject:    fmt.Sprintf("Policy Update - %s", policyType),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send policy update notification:", err)
		return err
	}
	log.Printf("Policy update notification sent to %v", email)
	return nil
}

// SendShopSecurityAlertNotification sends security-related alerts
func (e *emailService) SendShopSecurityAlertNotification(email, shopName, alertType, details string) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Security Alert: We detected a security event related to your shop account.</p>
		
		<p><strong>Alert Type:</strong> %s<br>
		<strong>Details:</strong> %s</p>
		
		<p>If this was not you, please secure your account immediately by changing your password.</p>
	`, shopName, alertType, details)

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Security Alert", content),
		Subject:    fmt.Sprintf("Security Alert - %s", alertType),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send security alert notification:", err)
		return err
	}
	log.Printf("Security alert notification sent to %v", email)
	return nil
}

// SendShopSubscriptionReminderNotification sends subscription/fee reminders
func (e *emailService) SendShopSubscriptionReminderNotification(email, shopName string, dueDate time.Time, amount float64) error {
	content := fmt.Sprintf(`
		<p>Dear <span class="highlight">%s</span>,</p>
		
		<p>Payment reminder: Your subscription payment is due soon.</p>
		
		<p><strong>Amount Due:</strong> ₦%.2f<br>
		<strong>Due Date:</strong> %s</p>
		
		<p>Please ensure your payment method is up to date to avoid service interruption.</p>
	`, shopName, amount, dueDate.Format("January 2, 2006"))

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body:       e.createEmailTemplate("Payment Reminder", content),
		Subject:    fmt.Sprintf("Payment Reminder - ₦%.2f Due %s", amount, dueDate.Format("Jan 2")),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send subscription reminder notification:", err)
		return err
	}
	log.Printf("Subscription reminder notification sent to %v", email)
	return nil
}
