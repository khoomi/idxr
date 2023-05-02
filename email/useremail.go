package email

import (
	"fmt"
	"log"
)

func SendWelcomeEmail(email, loginName string) {
	from := SetHeader{
		field: "From",
		value: []string{"no-reply@khoomi.com"},
	}

	to := SetHeader{
		field: "To",
		value: []string{email},
	}

	subject := SetHeader{
		field: "Subject",
		value: []string{"Welcome to Khoomi - Connecting Nigerian Creatives to the World!"},
	}

	body := SetBody{
		contentType: "text/html",
		body:        fmt.Sprintf("<body>\n    <h1>Welcome to Khoomi - Connecting Nigerian Creatives to the World!</h1>\n    <p>Dear %v,</p>\n    <p>\n      We are thrilled to welcome you to Khoomi, the new e-commerce platform for\n      Nigerian creatives. Thank you for joining our community of talented\n      artisans, small business owners, and passionate shoppers.\n    </p>\n    <p>\n      At Khoomi, our mission is to connect Nigerian creatives with a wider\n      audience, helping you to showcase your products and services to a broader\n      market. Our platform offers a user-friendly interface that allows you to\n      easily list your products and reach customers across Nigeria and beyond.\n    </p>\n    <p>\n      As a new user, you can set up your seller account and start listing your\n      products in just a few simple steps. Our team is always here to help you if\n      you have any questions or concerns, so please don't hesitate to reach out.\n    </p>\n    <p>\n      We are excited to have you on board and can't wait to see the amazing\n      products and services you have to offer. Thank you for choosing Khoomi!\n    </p>\n    <p>Best,</p>\n    <p>The Khoomi Team</p>\n  </body>", loginName),
	}

	compose := KhoomiEmailComposer{
		Header:        []SetHeader{from, to, subject},
		AddressHeader: nil,
		Body:          body,
		Attach:        "",
	}
	service := NewKhoomiEmailService(compose)
	err := service.SendMail()
	if err != nil {
		log.Println(err)
	}
}

func SendPasswordResetEmail(email, loginName, link string) {
	from := SetHeader{
		field: "From",
		value: []string{"no-reply@khoomi.com"},
	}

	to := SetHeader{
		field: "To",
		value: []string{email},
	}

	subject := SetHeader{
		field: "Subject",
		value: []string{"Khoomi Password Reset Request"},
	}

	body := SetBody{
		contentType: "text/html",
		body:        fmt.Sprintf("<body>\n    <h1>Khoomi Password Reset Request</h1>\n    <p>Dear %v,</p>\n    <p>\n      We received a request to reset the password for your Khoomi account. To\n      reset your password, please click the button below:\n    </p>\n    <div style=\"text-align: center;\">\n      <a\n        href=\"%v\"\n        style=\"\n          background-color: #FF5810;\n          color: #ffffff;\n          border-radius: 30px;\n          display: inline-block;\n          font-size: 16px;\n          font-weight: bold;\n          padding: 10px 16px;\n          text-align: center;\n          text-decoration: none;\n        \"\n        >Reset Password</a\n      >\n    </div>\n    <p>\n      If you did not request a password reset, please ignore this message and\n      your password will remain unchanged.\n    </p>\n    <p>Thank you,</p>\n    <p>The Khoomi Team</p>\n  </body>", loginName, link),
	}

	compose := KhoomiEmailComposer{
		Header:        []SetHeader{from, to, subject},
		AddressHeader: nil,
		Body:          body,
		Attach:        "",
	}
	service := NewKhoomiEmailService(compose)
	err := service.SendMail()
	if err != nil {
		log.Println(err)
	}
}

func SendPasswordResetSuccessfulEmail(email, loginName string) {
	from := SetHeader{
		field: "From",
		value: []string{"no-reply@khoomi.com"},
	}

	to := SetHeader{
		field: "To",
		value: []string{email},
	}

	subject := SetHeader{
		field: "Subject",
		value: []string{"Password Reset Successful"},
	}

	body := SetBody{
		contentType: "text/html",
		body:        fmt.Sprintf("<body>\n        <p>Dear %v,</p>\n        <p>Your password has been successfully reset.</p>\n        <p>If you did not request this password reset, please contact us immediately.</p>\n        <p>Best regards,</p>\n        <p>Your Application Team</p>\n    </body>", loginName),
	}

	compose := KhoomiEmailComposer{
		Header:        []SetHeader{from, to, subject},
		AddressHeader: nil,
		Body:          body,
		Attach:        "",
	}
	service := NewKhoomiEmailService(compose)
	err := service.SendMail()
	if err != nil {
		log.Println(err)
	}
}

func SendNewIpLoginNotification(email, loginName, ipaddr, loginTime string) {
	from := SetHeader{
		field: "From",
		value: []string{"no-reply@khoomi.com"},
	}

	to := SetHeader{
		field: "To",
		value: []string{email},
	}

	subject := SetHeader{
		field: "Subject",
		value: []string{"New IP Address Login Notification"},
	}

	body := SetBody{
		contentType: "text/html",
		body:        fmt.Sprintf(" <body>\n        <p>Dear %v,</p>\n        <p>This is to inform you that a new IP address has been used to log in to your account at %v.</p>\n        <p>IP Address: %v</p>\n         <p>If you did not log in from this location, please contact us or change your password immediately.</p>\n        <p>Best regards,</p>\n        <p>Your Application Team</p>\n    </body>", loginName, loginTime, ipaddr),
	}

	compose := KhoomiEmailComposer{
		Header:        []SetHeader{from, to, subject},
		AddressHeader: nil,
		Body:          body,
		Attach:        "",
	}
	service := NewKhoomiEmailService(compose)
	err := service.SendMail()
	if err != nil {
		log.Println(err)
	}
}
