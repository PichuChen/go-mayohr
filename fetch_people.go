package mayohr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type PersonData struct {
	Content              string `json:"Content"`
	PersonalPictureUrl   string `json:"PersonalPictureUrl"`
	EmployeeId           string `json:"EmployeeId"`
	EmployeeNumber       string `json:"EmployeeNumber"`
	JobTitle             string `json:"JobTitle"`
	ChineseName          string `json:"ChineseName"`
	EnglishName          string `json:"EnglishName"`
	ExtensionNumber      string `json:"ExtensionNumber"`
	BusinessEmail        string `json:"BusinessEmail"`
	BusinessMobileNumber string `json:"BusinessMobileNumber"`
	DepartmentId         string `json:"DepartmentId"`
	DeptCode             string `json:"DeptCode"`
	DeptName             string `json:"DeptName"`
	SupervisorId         string `json:"SupervisorId"`
	SupervisorCName      string `json:"SupervisorCName"`
	SupervisorEName      string `json:"SupervisorEName"`
}

func (c *Client) PeopleSearch(keyword string, pageNumber int, pageSize int) (data []PersonData, totalCount int, err error) {
	// URL: https://apolloxe.mayohr.com/backend/fd/api/employees/peoplesearch
	// Method: GET
	// Header:
	// 	Authorization: {{id_token}}
	// Query:
	// 	Keyword={{keyword}}&PageNumber={{pageNumber}}&PageSize={{pageSize}}
	// Response:
	// {"Meta":{"PageNumber":3,"PageSize":10,"PageCount":64,"HttpStatusCode":"200"},
	// 	"Data":[
	// 		{
	// 			"Content":"","PersonalPictureUrl":null,"EmployeeId":"",
	// 			"EmployeeNumber":"","JobTitle":"","ChineseName":"","EnglishName":"",
	// 			"ExtensionNumber":null,"BusinessEmail":"","BusinessMobileNumber":null,
	// 			"DepartmentId":"","DeptCode":"","DeptName":"",
	// 			"SupervisorId":"","SupervisorCName":"","SupervisorEName":""
	// 		}
	// ]}

	args := url.Values{
		"Keyword":    {keyword},
		"PageNumber": {fmt.Sprintf("%d", pageNumber)},
		"PageSize":   {fmt.Sprintf("%d", pageSize)},
	}
	requestURL := "https://apolloxe.mayohr.com/backend/fd/api/employees/peoplesearch?" + args.Encode()
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", c.IDToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	var respPayload struct {
		Meta struct {
			PageNumber     int    `json:"PageNumber"`
			PageSize       int    `json:"PageSize"`
			PageCount      int    `json:"PageCount"`
			HttpStatusCode string `json:"HttpStatusCode"`
		} `json:"Meta"`
		Data []PersonData `json:"Data"`
	}
	respByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read response body: %v", err)
	}
	// fmt.Printf("resp.Body: %v\n", string(respByte))

	err = json.Unmarshal(respByte, &respPayload)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return nil, 0, fmt.Errorf("failed to decode response body: %v", err)
	}

	// fmt.Printf("respPayload: %v\n", respPayload)
	totalCount = respPayload.Meta.PageCount

	return respPayload.Data, totalCount, nil

}

func (c *Client) FetchAllPeople() (people []PersonData, err error) {
	_, totalCount, err := c.PeopleSearch("0", 0, 10)
	if err != nil {
		fmt.Printf("failed to get people: %v", err)
		return
	}

	totalPeopleList := make([]PersonData, totalCount)
	wg := sync.WaitGroup{}
	for i := 0; i < totalCount/10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			people, _, err := c.PeopleSearch("0", i, 10)
			if err != nil {
				fmt.Printf("failed to get people: %v", err)
				return
			}
			for j, person := range people {
				totalPeopleList[i*10+j] = person
			}
		}(i)
	}
	wg.Wait()
	return totalPeopleList, nil
}
