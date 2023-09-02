package mayohr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

type Client struct {
	Username string
	Password string

	cookieRVT string
	cookieSD  string
	htmlRVT   string

	AccessToken  string
	RefreshToken string
	Code         string
	IDToken      string
}

// NewClient returns a new Client given a username and password.
func NewClient(username, password string) *Client {
	return &Client{
		Username: username,
		Password: password,
	}
}

func (c *Client) Login() (err error) {
	err = c.getRequestVerificationToken()
	if err != nil {
		return fmt.Errorf("failed to get request verification token: %v", err)
	}
	err = c.getAccessTokenAndRefreshToken()
	if err != nil {
		return fmt.Errorf("failed to get access token and refresh token: %v", err)
	}
	err = c.getIDToken()
	if err != nil {
		return fmt.Errorf("failed to get id token: %v", err)
	}
	return nil
}

func (c *Client) getRequestVerificationToken() error {
	// url: https://auth.mayohr.com/HRM/Account/Login
	// method: GET
	// response (from cookie):
	// 	__RequestVerificationToken={{cookie_rvt}}
	// 	_sd={{cookie_sd}}
	// html:
	// 	<input name="__RequestVerificationToken" type="hidden" value="{{html_rvt}}" />

	req, err := http.NewRequest("GET", "https://auth.mayohr.com/HRM/Account/Login", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		fmt.Printf("cookie: %v\n", cookie)
		if cookie.Name == "__RequestVerificationToken" {
			c.cookieRVT = cookie.Value
		} else if cookie.Name == "_sd" {
			c.cookieSD = cookie.Value
		}
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// fmt.Printf("resp.Body: %v\n", string(respByte))

	regex := regexp.MustCompile(`<input name="__RequestVerificationToken" type="hidden" value="(.*)" />`)
	matches := regex.FindStringSubmatch(string(respByte))
	if len(matches) != 2 {
		return fmt.Errorf("failed to find __RequestVerificationToken")
	}
	c.htmlRVT = matches[1]
	return nil
}

func (c *Client) getAccessTokenAndRefreshToken() error {
	// url: https://auth.mayohr.com/token
	// method: POST
	// body: grant_type=password&userName={{username}}&password={{password}}&__RequestVerificationToken={{html_rvt}}
	// header: Cookie: __RequestVerificationToken={{cookie_rvt}}; _sd={{cookie_sd}}
	// success response: {
	// 	"access_token": "{{access_token}}",
	// 	"refresh_token": "{{refresh_token}}",
	// 	"code": "{{code}}",
	// }
	// error response: {
	// 	"error": "{{error_code}}",
	// 	"error_description": "{{error_description}}"
	// }

	reqPayload := url.Values{
		"grant_type":                 {"password"},
		"userName":                   {c.Username},
		"password":                   {c.Password},
		"__RequestVerificationToken": {c.htmlRVT},
	}
	// urlencoding
	reqByte := []byte(reqPayload.Encode())

	req, err := http.NewRequest("POST", "https://auth.mayohr.com/Token", bytes.NewBuffer(reqByte))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("__RequestVerificationToken=%s; _sd=%s", c.cookieRVT, c.cookieSD))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var respPayload map[string]interface{}
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// fmt.Printf("resp.Body: %v\n", string(respByte))

	err = json.Unmarshal(respByte, &respPayload)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return fmt.Errorf("failed to decode response body: %v", err)
	}

	c.AccessToken, _ = respPayload["access_token"].(string)
	c.RefreshToken, _ = respPayload["refresh_token"].(string)
	c.Code, _ = respPayload["code"].(string)

	// debug
	// fmt.Printf("respPayload: %v\n", respPayload)
	if c.AccessToken == "" {
		return fmt.Errorf("failed to get access token: %v", respPayload["error_description"])
	}

	return nil
}

func (c *Client) getIDToken() error {
	// Thanks to https://hackmd.io/@TmvqCeJ1SS6VT0toHIHQWQ/r1r7o5d08

	// https://authcommon.mayohr.com/api/auth/checkticket
	// method: GET
	// args: code={{code}}&response_type=id_token
	// response: {
	// 	"id_token": "{{id_token}}"
	// }

	reqArgs := url.Values{
		"code":          {c.Code},
		"response_type": {"id_token"},
	}
	reqURL := fmt.Sprintf("https://authcommon.mayohr.com/api/auth/checkticket?%s", reqArgs.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var respPayload map[string]interface{}
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	// fmt.Printf("resp.Body: %v\n", string(respByte))

	err = json.Unmarshal(respByte, &respPayload)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return fmt.Errorf("failed to decode response body: %v", err)
	}
	c.IDToken, _ = respPayload["id_token"].(string)

	return nil
}
