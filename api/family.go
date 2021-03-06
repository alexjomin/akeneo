package akeneo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

type Family struct {
	Code                  string              `json:"code"`
	AttributeAsLabel      string              `json:"attribute_as_label"`
	AttributeAsImage      string              `json:"attribute_as_image,omitempty"`
	Attributes            []string            `json:"attributes,omitempty"`
	AttributeRequirements map[string][]string `json:"attribute_requirements,omitempty"`
	Labels                map[string]string   `json:"labels,omitempty"`
}

type FamilyApi ApiService

type FamilyItem struct {
	Family
	ResponseLinks `json:"_links"`
}

type FamiliesResponse struct {
	Response
	Data struct {
		Items []FamilyItem `json:"items"`
	} `json:"_embedded"`
}

func (service *FamilyApi) GetAll(opts RequestOpts) (*FamiliesResponse, *ApiError) {
	headers := service.client.getHeadersForRequest()
	queryParams := &url.Values{}

	for _, key := range []string{"page", "limit", "withCount"} {
		if value, ok := opts[key].(string); ok {
			queryParams.Add(key, value)
		}
	}

	response, err := service.client.DoRequest("GET", "families", headers, nil, queryParams)

	if err != nil {
		return nil, &ApiError{Message: err.Error()}
	}

	defer response.Body.Close()
	if response.StatusCode >= 300 {
		msg, _ := ioutil.ReadAll(response.Body)
		return nil, &ApiError{Code: response.StatusCode, Status: response.Status, Message: fmt.Sprintf("%s", msg)}
	}

	resp := &FamiliesResponse{}

	if err = json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, &ApiError{Message: err.Error()}
	}

	return resp, nil
}

func (service *FamilyApi) Get(code string) (*Family, *ApiError) {
	headers := service.client.getHeadersForRequest()
	uri := fmt.Sprintf("families/%s", code)

	response, err := service.client.DoRequest("GET", uri, headers, nil, nil)
	if err != nil {
		return nil, &ApiError{Message: err.Error()}
	}

	defer response.Body.Close()

	if response.StatusCode >= 300 {
		msg, _ := ioutil.ReadAll(response.Body)
		return nil, &ApiError{Code: response.StatusCode, Status: response.Status, Message: fmt.Sprintf("%s", msg)}
	}

	var family = &Family{}
	if err = json.NewDecoder(response.Body).Decode(&family); err != nil {
		return nil, &ApiError{Message: err.Error()}
	}

	return family, nil
}

func (service *FamilyApi) Create(family *Family) *ApiError {
	headers := service.client.getHeadersForRequest()
	body, _ := json.Marshal(family)

	response, err := service.client.DoRequest("POST", "families", headers, body, nil)
	if err != nil {
		return &ApiError{Message: err.Error()}
	}

	defer response.Body.Close()

	if response.StatusCode >= 400 {
		msg, _ := ioutil.ReadAll(response.Body)
		return &ApiError{Code: response.StatusCode, Status: response.Status, Message: fmt.Sprintf("%s", msg)}
	}

	return nil
}

func (service *FamilyApi) Upsert(family *Family) *ApiError {

	headers := service.client.getHeadersForRequest()
	uri := fmt.Sprintf("families/%s", family.Code)
	body, _ := json.Marshal(family)

	response, err := service.client.DoRequest("PATCH", uri, headers, body, nil)
	if err != nil {
		return &ApiError{Code: 0, Message: err.Error()}
	}

	defer response.Body.Close()

	if response.StatusCode >= 300 {
		msg, _ := ioutil.ReadAll(response.Body)
		return &ApiError{Code: response.StatusCode, Status: response.Status, Message: fmt.Sprintf("%s", msg)}
	}

	return nil
}

func (service *FamilyApi) BatchUpsert(families []*Family) ([]*ResponseBody, *ApiError) {
	headers := service.client.getHeadersForBatchRequest()
	var body []byte

	for _, bodyItem := range families {
		bodyItem, _ := json.Marshal(bodyItem)
		body = append(body, bodyItem...)
		body = append(body, '\n')
	}

	response, err := service.client.DoRequest("PATCH", "families", headers, body, nil)
	if err != nil {
		return nil, &ApiError{Message: err.Error()}
	}

	defer response.Body.Close()

	if response.StatusCode >= 300 {
		msg, _ := ioutil.ReadAll(response.Body)
		return nil, &ApiError{Code: response.StatusCode, Status: response.Status, Message: fmt.Sprintf("%s", msg)}
	}

	var apiResponse []*ResponseBody
	scanner := bufio.NewScanner(response.Body)

	for scanner.Scan() {
		var responseLine *ResponseBody
		var reader = bytes.NewReader(scanner.Bytes())
		if err = json.NewDecoder(reader).Decode(&responseLine); err != nil {
			return nil, &ApiError{Message: err.Error()}
		}

		apiResponse = append(apiResponse, responseLine)
	}

	return apiResponse, nil
}
