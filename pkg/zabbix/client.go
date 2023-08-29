package zabbix

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type Client struct {
	url      string
	username string
	password string
}

type Request struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      int         `json:"id"`
	Auth    string      `json:"auth,omitempty"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *Error      `json:"error"`
	Id      int         `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func NewClient(url, username, password string) (*Client, error) {
	if url == "" || username == "" || password == "" {
		return nil, errors.New("url, username, and password are required")
	}
	return &Client{url, username, password}, nil
}

func (c *Client) Do(request Request) (*Response, error) {
	// Marshal the request body
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the response body
	var response Response
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	// Check for errors in the response
	if response.Error != nil {
		return nil, errors.New(response.Error.Message)
	}

	return &response, nil
}
