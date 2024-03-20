package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"khoomi-api-io/khoomi_api/config"
	"log"
	"net/http"
)

type KhoomiEmailComposer struct {
	Body       string
	Subject    string
	Sender     string
	SenderName string
	To         string
	ToName     string
}

func SendMail(mail KhoomiEmailComposer) error {
	data := map[string]interface{}{
		"sender": map[string]string{
			"name":  mail.SenderName,
			"email": mail.Sender,
		},
		"to": []map[string]string{
			{
				"email": mail.To,
				"name":  mail.ToName,
			},
		},
		"subject":     mail.Subject,
		"htmlContent": mail.Body,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Email service: Error marshaling JSON:", err)
		return err
	}

	MailApiKey := config.LoadEnvFor("MAIL_API_KEY")
	MailEndPoint := config.LoadEnvFor("MAIL_ENDPOINT")
	req, err := http.NewRequest("POST", MailEndPoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Email service:", err)
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", MailApiKey)
	req.Header.Set("content-type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	responseBody := &bytes.Buffer{}
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return err
	}

	fmt.Println(responseBody.String())

	return nil
}
