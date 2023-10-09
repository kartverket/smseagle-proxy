package notifier

import (
	"fmt"
	"net/http"
)

func sendSMS(baseUrl string, accessToken string, phoneNumber string, message string) error {
	requestUrl := fmt.Sprintf("%s/http_api/send_sms?access_token=%s&to=%s&message=%s", baseUrl, accessToken, phoneNumber, message)
	res, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	return nil
}

func call(baseUrl string, accessToken string, phoneNumber string) error {
	requestUrl := fmt.Sprintf("%s/http_api/call_with_termination?access_token=%s&to=%s", baseUrl, accessToken, phoneNumber)
	res, err := http.Get(requestUrl)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
	}

	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", res.StatusCode)
	return nil
}
