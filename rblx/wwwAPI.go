package rblx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SparklyCatTF2/Reaper/globals"
)

type SellItem struct {
	AssetID     int    `json:"assetId"`
	UserAssetID int    `json:"UserAssetId"`
	Price       int    `json:"price"`
	Sell        string `json:"sell"`
}

type SellItemResponse struct {
	Sold  bool   `json:"isValid"`
	Error string `json:"error"`
}

// GetResellers fetches the 10 cheapest resellers for an item
func (session *RBLXSession) SellItem(assetID int64, userAssetID int64, price int64) (*SellItemResponse, *Error) {
	query_string := url.Values{}
	query_string.Add("assetId", string(assetID))
	query_string.Add("userAssetId", string(userAssetID))
	query_string.Add("price", string(price))
	query_string.Add("sell", "true")

	sellItemReq, _ := http.NewRequest("POST", "https://www.roblox.com/asset/toggle-sale", strings.NewReader(query_string.Encode()))
	sellItemReq.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", globals.Config.Cookie))
	sellItemReq.Header.Add("X-CSRF-TOKEN", *session.XCSRFToken)
	sellItemReq.Header.Add("Content-Type", "application/json")
	resp, respError := session.Client.Do(sellItemReq)
	if respError != nil {
		return nil, NewCustomError(respError, -1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, StatusCodeToError(resp.StatusCode)
	}

	var sellItemData SellItemResponse
	json.NewDecoder(resp.Body).Decode(&sellItemData)

	if sellItemData.Error != "" {
		return nil, NewCustomError(errors.New(sellItemData.Error), -1)
	}

	return &sellItemData, nil
}