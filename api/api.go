package api

import (
	"encoding/json"
	"io"
	"net/http"
)

var apiUrl string

func Initialize(api string) {
	apiUrl = api
}

func GetApiResponse[T any](path string) (*T, error) {
	response, err := http.Get(apiUrl + path)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var r ApiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	resultBytes, err := json.Marshal(r.Result)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
