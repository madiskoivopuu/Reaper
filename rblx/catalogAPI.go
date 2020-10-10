package rblx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var counter = 0

// Details endpoint post data & response structs
type Item struct {
	ID       int64  `json:"id"`
	ItemType string `json:"itemType"` // const
}
type DetailsPost struct {
	Items []Item `json:"items"`
}
type DetailsResponse struct {
	Data []struct {
		ID          int64 `json:"id"`
		LowestPrice int64 `json:"lowestPrice"`
	} `json:"data"`
}

// Recommendations endpoint post data & response structs
type RecommendationsResponse struct {
	Data []struct {
		Item struct {
			Price int `json:"price"`
		} `json:"item"`
	} `json:"data"`
}

// GetCatalogDetails fetches the catalog details for asset ids
func (session *RBLXSession) GetCatalogDetails(assetIDs []int64) (*DetailsResponse, *Error) {
	// format the assetIds
	itemDetails := DetailsPost{}
	for _, assetID := range assetIDs {
		item := Item{ID: assetID, ItemType: "Asset"}
		itemDetails.Items = append(itemDetails.Items, item)
	}
	postJSON, jsonError := json.Marshal(itemDetails)
	if jsonError != nil {
		return nil, NewCustomError(jsonError, -1)
	}

	req, _ := http.NewRequest("POST", "https://catalog.roblox.com/v1/catalog/items/details", bytes.NewReader(postJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-CSRF-TOKEN", *session.XCSRFToken)
	req.Header.Add("Cookie", ".ROBLOSECURITY="+session.Cookie)
	resp, respError := session.Client.Do(req)
	if respError != nil {
		return nil, NewCustomError(respError, -1)
	}
	defer resp.Body.Close()

	// error checks
	if resp.StatusCode != 200 {
		return nil, StatusCodeToError(resp.StatusCode)
	}

	respText, readError := ioutil.ReadAll(resp.Body)
	if readError != nil {
		return nil, NewCustomError(readError, -1)
	}

	var details DetailsResponse
	jsonParseError := json.Unmarshal(respText, &details)
	//fmt.Println(string(respText))

	if jsonParseError != nil {
		return nil, NewCustomError(jsonParseError, -1)
	}


	return &details, nil
}

// GetRecommendations fetches X number of recommendations for a certain asset id (with some sorf of context asset id)
func (session *RBLXSession) GetRecommendations(assetID int64, contextAssetID int64, numItems int64) (*RecommendationsResponse, *Error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://catalog.roblox.com/v1/recommendations/asset/%d?contextAssetId=%d&numItems=%d", assetID, contextAssetID, numItems), nil)
	req.Header.Add("Cookie", ".ROBLOSECURITY="+session.Cookie)
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

	var recommendations RecommendationsResponse
	jsonParseError := json.Unmarshal(respText, &recommendations)
	if jsonParseError != nil {
		return nil, NewCustomError(jsonParseError, -1)
	}

	return &recommendations, nil
}
