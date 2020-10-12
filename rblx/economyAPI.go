package rblx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SparklyCatTF2/Reaper/globals"
)

type ResellersResponse struct {
	Data []struct {
		UserAssetID int64 `json:"userAssetId"`
		Seller      struct {
			ID int64 `json:"id"`
		} `json:"seller"`
		Price int64 `json:"price"`
	} `json:"data"`
}

type PurchasePost struct {
	AssetID          int64 `json:"-"`
	ExpectedCurrency int64 `json:"expectedCurrency"`
	ExpectedPrice    int64 `json:"expectedPrice"`
	ExpectedSellerID int64 `json:"expectedSellerId"`
	UserAssetID      int64 `json:"userAssetId"`
}
type PurchaseResponse struct {
	Purchased          bool   `json:"purchased"`
	Reason             string `json:"reason"`
	Price              int    `json:"price"`
	AssetID            int    `json:"assetId"`
	AssetName          string `json:"assetName"`
	SellerName         string `json:"sellerName"`
}

// GetResellers fetches the 10 cheapest resellers for an item
func (session *RBLXSession) GetResellers(assetID int64) (*ResellersResponse, *Error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://economy.roblox.com/v1/assets/%d/resellers?limit=10", assetID), nil)
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

	var resellers ResellersResponse
	jsonParseError := json.Unmarshal(respText, &resellers)
	if jsonParseError != nil {
		return nil, NewCustomError(jsonParseError, -1)
	}

	return &resellers, nil
}

func (session *RBLXSession) PurchaseItem(assetId int64, purchaseStruct PurchasePost) (*PurchaseResponse, *Error) {
	jsonData, jsonError := json.Marshal(purchaseStruct)
	if jsonError != nil {
		return nil, NewCustomError(jsonError, -1)
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://economy.roblox.com/v1/purchases/products/%d", globals.CachedProductIDs[assetId]), bytes.NewReader(jsonData))
	req.Header.Add("Cookie", ".ROBLOSECURITY="+session.Cookie)
	req.Header.Add("X-CSRF-TOKEN", *session.XCSRFToken)
	req.Header.Add("Content-Type", "application/json")
	resp, respError := session.Client.Do(req)
	if respError != nil {
		return nil, NewCustomError(respError, -1)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, StatusCodeToError(resp.StatusCode)
	}

	respText, readAllError := ioutil.ReadAll(resp.Body)
	if readAllError != nil {
		return nil, NewCustomError(readAllError, -1)
	}

	var purchaseResponse PurchaseResponse
	jsonParseError := json.Unmarshal(respText, &purchaseResponse)
	if jsonParseError != nil {
		return nil, NewCustomError(jsonParseError, -1)
	}

	return &purchaseResponse, nil
}
