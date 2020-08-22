package rblx

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ResellersResponse struct {
	Data []struct {
		UserAssetID int `json:"userAssetId"`
		Seller      struct {
			ID int `json:"id"`
		} `json:"seller"`
		Price int `json:"price"`
	} `json:"data"`
}

// GetResellers fetches the 10 cheapest resellers for an item
func (session *RBLXSession) GetResellers(assetID int64) (*ResellersResponse, *Error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://economy.roblox.com/v1/assets/%d/resellers?limit=10", assetID), nil)
	resp, respError := session.Client.Do(req)
	if respError != nil {
		return nil, NewCustomError(respError, -1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, StatusCodeToError(resp.StatusCode)
	}

	respText, readError := ioutil.ReadAll(resp.Body)
	if readError != nil {
		return nil, NewCustomError(readError, -1)
	}

	var resellers ResellersResponse
	jsonParseError := json.Unmarshal(respText, &resellers)
	if jsonParseError != nil {
		return nil, NewCustomError(jsonParseError, -1)
	}

	return &resellers, nil
}
