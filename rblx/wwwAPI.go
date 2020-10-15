package rblx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SparklyCatTF2/Reaper/globals"
	"net/http"
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

// Sells item
func (session *RBLXSession) SellItem(assetID int64, userAssetID int64, price int64) (*SellItemResponse, *Error) {
	var sellitemjson SellItem
	sellitemjson = SellItem{AssetID: int(assetID), UserAssetID: int(userAssetID), Price: int(price), Sell: "true"}
	sellitem, err := json.Marshal(sellitemjson)
	if err != nil {
		fmt.Println("Something went wrong building JSON to sell item")
	}

	sellItemReq, _ := http.NewRequest("POST", "https://www.roblox.com/asset/toggle-sale", bytes.NewBuffer(sellitem))
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