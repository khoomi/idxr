package email

import (
	"fmt"
	"khoomi-api-io/api/pkg/util"
	"log"
)

// SendNewListingEmail sends a notification email to a user when a new listing is successfully posted in their shop on Khoomi.
func SendNewListingEmail(email, sellerName, listingTitle string) {
	mail := util.KhoomiEmailComposer{
		To:         email,
		ToName:     sellerName,
		Sender:     "no-reply@khoomi.com",
		SenderName: "Khoomi Online",
		Body:       fmt.Sprintf("<html><head><style>body{font-family: Arial, sans-serif; font-size:14px;}p{margin-bottom: 10px;}.highlight{font-weight: bold;color: #FF5810;}</style></head><body><p>Dear <span class=\"highlight\">%v</span>,</p><p>Congratulations on creating a new listing on Khoomi!</p><p>Your listing titled <span class=\"highlight\">'%v'</span> has been successfully created and is now live on our platform. This means that your products and services are now accessible to a wide audience of passionate shoppers.</p><p>We wish you the best of luck with your new listing. If you have any questions or need assistance, please feel free to reach out to our support team.</p><p>Thank you for choosing Khoomi as your e-commerce partner. We look forward to seeing your business thrive on our platform.</p><p>Best,</p><p>The Khoomi Team</p></body></html>", sellerName, listingTitle),
		Subject:    fmt.Sprintf("Congratulations on Your New Listing - %v", listingTitle),
	}

	err := util.SendMail(mail)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("New listing email sent to %v", email)
	}
}
