package mayohr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type BasicInfo struct {
	PersonalId      string `json:"PersonalId"`
	LastName        string `json:"LastName"`
	FirstName       string `json:"FirstName"`
	ChineseName     string `json:"ChineseName"`
	EnglishName     string `json:"EnglishName"`
	Gender          string `json:"Gender"`
	GenderText      string `json:"GenderText"`
	Birthday        string `json:"Birthday"`
	Nation          string `json:"Nation"`
	NationText      string `json:"NationText"`
	PersonalPicture string `json:"PersonalPicture"`
	CreateOn        string `json:"CreateOn"`
	Creater         string `json:"Creater"`
	UpdateOn        string `json:"UpdateOn"`
	Updater         string `json:"Updater"`
}

func (c *Client) GetBasicInfo() (*BasicInfo, error) {
	// URL: https://apolloxe.mayohr.com/backend/fd/api/users/basicinfos
	// Method: GET
	// Header:
	//   Authorization: {{id_token}}
	// Response:
	// {"Meta":{"HttpStatusCode":"200"},"Data":{"PersonalId":"{{UUID}}","LastName":"","FirstName":"",
	// "ChineseName":"","EnglishName":"Pichu","Gender":"","GenderText":"","Birthday":"2006-01-02T15:04:05+08:00","Nation":"",
	// "NationText":"","ArmyStatus":null,"ArmyStatusText":"","ArmyType":"0","ArmyTypeText":"","ArmyStart":null,"ArmyEnd":null,"ExemptReason":null,
	// "MaritalStatus":"0","MaritalStatusText":"","EntryTime":null,"IDType":1,"IDTypeText":"","IDNumber":"{{IDNumber}}","IDExpiryDate":null,
	// "IDType2":"","IDNumber2":"","IDExpiryDate2":null,"IDType3":"","IDNumber3":"","IDExpiryDate3":null,"PersonalPicture":"{{UUID}}",
	// "CreateOn":"2006-01-02T15:04:05+08:00","Creater":"","UpdateOn":"2006-01-02T15:04:05+08:00","Updater":""}}

	req, err := http.NewRequest("GET", "https://apolloxe.mayohr.com/backend/fd/api/users/basicinfos", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", c.IDToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var respPayload struct {
		Meta struct {
			HttpStatusCode string `json:"HttpStatusCode"`
		} `json:"Meta"`
		Data BasicInfo `json:"Data"`
	}

	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	// fmt.Printf("resp.Body: %v\n", string(respByte))
	err = json.Unmarshal(respByte, &respPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	if respPayload.Meta.HttpStatusCode != "200" {
		return nil, fmt.Errorf("failed to get basic info: %v", respPayload)
	}

	return &respPayload.Data, nil

}
