package services

import (
	"fmt"
	"khoomi-api-io/api/pkg/util"
	"log"
	"strings"
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

// SendShopNewOrderNotification sends an email notification when a new order is received
func (e *emailService) SendShopNewOrderNotification(email, shopName, orderID, customerName string, orderTotal float64) error {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.amount{font-size: 18px; font-weight: bold; color: #28a745;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>Great news! You have received a new order on Khoomi!</p>
				<p><strong>Order Details:</strong></p>
				<p>Order ID: <span class="highlight">%v</span></p>
				<p>Customer: %v</p>
				<p>Order Total: <span class="amount">₦%.2f</span></p>
				<p>Please log in to your seller dashboard to process this order promptly.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, orderID, customerName, orderTotal),
		Subject: fmt.Sprintf("New Order Received - Order #%v", orderID),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.amount{font-size: 18px; font-weight: bold; color: #28a745;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>Payment has been successfully confirmed for your order!</p>
				<p><strong>Payment Details:</strong></p>
				<p>Order ID: <span class="highlight">%v</span></p>
				<p>Amount Received: <span class="amount">₦%.2f</span></p>
				<p>You can now proceed with fulfilling this order.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, orderID, amount),
		Subject: fmt.Sprintf("Payment Confirmed - Order #%v", orderID),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.warning{color: #dc3545; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="warning">Payment failed</span> for an order in your shop.</p>
				<p><strong>Order Details:</strong></p>
				<p>Order ID: <span class="highlight">%v</span></p>
				<p>Failure Reason: %v</p>
				<p>Please check your seller dashboard for more details and follow up if necessary.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, orderID, reason),
		Subject: fmt.Sprintf("Payment Failed - Order #%v", orderID),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.warning{color: #dc3545;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>An order has been <span class="warning">cancelled</span> in your shop.</p>
				<p><strong>Order Details:</strong></p>
				<p>Order ID: <span class="highlight">%v</span></p>
				<p>Customer: %v</p>
				<p>Please check your seller dashboard for more details about this cancellation.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, orderID, customerName),
		Subject: fmt.Sprintf("Order Cancelled - Order #%v", orderID),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.warning{color: #ffc107; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="warning">Low stock alert!</span> One of your products is running low on inventory.</p>
				<p><strong>Product Details:</strong></p>
				<p>Product: <span class="highlight">%v</span></p>
				<p>Current Stock: <strong>%d units</strong></p>
				<p>Threshold: %d units</p>
				<p>Consider restocking this product to avoid going out of stock.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, productName, currentStock, threshold),
		Subject: fmt.Sprintf("Low Stock Alert - %v", productName),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.danger{color: #dc3545; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="danger">Out of stock alert!</span> One of your products is now completely out of stock.</p>
				<p><strong>Product Details:</strong></p>
				<p>Product: <span class="highlight">%v</span></p>
				<p>Current Stock: <strong>0 units</strong></p>
				<p>This product is no longer available for purchase. Please restock as soon as possible.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, productName),
		Subject: fmt.Sprintf("Out of Stock Alert - %v", productName),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.success{color: #28a745; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="success">Inventory restocked!</span> Your product inventory has been updated.</p>
				<p><strong>Product Details:</strong></p>
				<p>Product: <span class="highlight">%v</span></p>
				<p>New Stock Level: <strong>%d units</strong></p>
				<p>Your product is now available for purchase again.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, productName, newStock),
		Subject: fmt.Sprintf("Inventory Restocked - %v", productName),
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
	stars := strings.Repeat("⭐", rating)
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.stars{font-size: 16px;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>You have received a new review on one of your products!</p>
				<p><strong>Review Details:</strong></p>
				<p>Product: <span class="highlight">%v</span></p>
				<p>Reviewer: %v</p>
				<p>Rating: <span class="stars">%v</span> (%d/5)</p>
				<p>Check your seller dashboard to read the full review and respond if needed.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, productName, reviewerName, stars, rating),
		Subject: fmt.Sprintf("New Review Received - %v", productName),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>You have received a new message from a customer!</p>
				<p><strong>Message Details:</strong></p>
				<p>From: <span class="highlight">%v</span></p>
				<p>Subject: %v</p>
				<p>Please log in to your seller dashboard to read and respond to this message.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, customerName, subject),
		Subject: fmt.Sprintf("New Customer Message - %v", subject),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.warning{color: #ffc107; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="warning">Return request received!</span> A customer has requested to return an order.</p>
				<p><strong>Return Details:</strong></p>
				<p>Order ID: <span class="highlight">%v</span></p>
				<p>Customer: %v</p>
				<p>Reason: %v</p>
				<p>Please review this return request in your seller dashboard and take appropriate action.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, orderID, customerName, reason),
		Subject: fmt.Sprintf("Return Request - Order #%v", orderID),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.amount{font-size: 18px; font-weight: bold; color: #28a745;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>Here's your sales summary for the %v:</p>
				<p><strong>Sales Performance:</strong></p>
				<p>Total Revenue: <span class="amount">₦%.2f</span></p>
				<p>Total Orders: <strong>%d orders</strong></p>
				<p>Keep up the great work! Check your seller dashboard for detailed analytics.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, period, totalSales, orderCount),
		Subject: fmt.Sprintf("Sales Summary - %v", period),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.milestone{font-size: 24px; font-weight: bold; color: #28a745;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>🎉 <strong>Congratulations!</strong> You've reached a new revenue milestone!</p>
				<p><span class="milestone">₦%.2f</span></p>
				<p>Achievement Period: %v</p>
				<p>This is a fantastic achievement! Keep growing your business on Khoomi.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, milestone, period),
		Subject: fmt.Sprintf("🎉 Revenue Milestone Reached - ₦%.2f!", milestone),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.trend{color: #28a745; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="trend">🔥 Trending Product Alert!</span></p>
				<p>Your product "<span class="highlight">%v</span>" is performing exceptionally well!</p>
				<p><strong>Performance Stats:</strong></p>
				<p>Sales in %v: <strong>%d units</strong></p>
				<p>Consider promoting this product more or creating similar items to capitalize on this trend.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, productName, period, salesCount),
		Subject: fmt.Sprintf("🔥 Trending Product - %v", productName),
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
	var statusColor, statusMessage string
	switch status {
	case "approved":
		statusColor = "#28a745"
		statusMessage = "Your shop has been successfully verified!"
	case "pending":
		statusColor = "#ffc107"
		statusMessage = "Your shop verification is under review."
	case "rejected":
		statusColor = "#dc3545"
		statusMessage = "Your shop verification was not approved."
	default:
		statusColor = "#6c757d"
		statusMessage = "Your shop verification status has been updated."
	}

	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.status{color: %v; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p>Your account verification status has been updated.</p>
				<p><span class="status">%v</span></p>
				<p>Status: <strong>%v</strong></p>
				<p>Check your seller dashboard for more details and next steps.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, statusColor, shopName, statusMessage, status),
		Subject: fmt.Sprintf("Account Verification Update - %v", shopName),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.info{color: #17a2b8; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="info">Policy Update Notice</span></p>
				<p>We have updated our <strong>%v</strong> policy.</p>
				<p><strong>Summary of Changes:</strong></p>
				<p>%v</p>
				<p>Please review the updated policy in your seller dashboard to ensure compliance.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, policyType, summary),
		Subject: fmt.Sprintf("Policy Update - %v", policyType),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.alert{color: #dc3545; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="alert">🚨 Security Alert</span></p>
				<p>We detected a security event related to your shop account.</p>
				<p><strong>Alert Type:</strong> %v</p>
				<p><strong>Details:</strong> %v</p>
				<p>If this was not you, please secure your account immediately by changing your password.</p>
				<p>Best regards,</p>
				<p>The Khoomi Security Team</p>
			</body>
		</html>`, shopName, alertType, details),
		Subject: fmt.Sprintf("🚨 Security Alert - %v", alertType),
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
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     shopName,
		Sender:     e.sender,
		SenderName: e.senderName,
		Body: fmt.Sprintf(`<html>
			<head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}.reminder{color: #ffc107; font-weight: bold;}.amount{font-size: 16px; font-weight: bold;}</style></head>
			<body>
				<p>Dear <span class="highlight">%v</span>,</p>
				<p><span class="reminder">💳 Payment Reminder</span></p>
				<p>Your subscription payment is due soon.</p>
				<p><strong>Payment Details:</strong></p>
				<p>Amount Due: <span class="amount">₦%.2f</span></p>
				<p>Due Date: <strong>%v</strong></p>
				<p>Please ensure your payment method is up to date to avoid service interruption.</p>
				<p>Best regards,</p>
				<p>The Khoomi Team</p>
			</body>
		</html>`, shopName, amount, dueDate.Format("January 2, 2006")),
		Subject: fmt.Sprintf("Payment Reminder - ₦%.2f Due %v", amount, dueDate.Format("Jan 2")),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println("Failed to send subscription reminder notification:", err)
		return err
	}
	log.Printf("Subscription reminder notification sent to %v", email)
	return nil
}
