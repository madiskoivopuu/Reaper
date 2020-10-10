package threading

import (
	"encoding/json"
	"fmt"
	"github.com/SparklyCatTF2/Reaper/globals"
	"net/http"
	"strconv"

	"github.com/SparklyCatTF2/Reaper/rblx"
)
var client = &http.Client{}

type GetItemInfo struct {
	Name string `json:"Name"`
}

// FetchResellers gets resellers for an asset ID and then passes the details through a channel
func FetchResellers(purchaseDetails *rblx.PurchasePost, snipeAccountSession *rblx.RBLXSession, fetchChannel chan *rblx.ResellersResponse) {

	//s1 := globals.GetTimeInMs()
	resellerDetails, err := snipeAccountSession.GetResellers(purchaseDetails.AssetID)
	//s2 := globals.GetTimeInMs() - s1
	//fmt.Println(s2)

	if err != nil {
		fmt.Println(err)

		switch err.Type {
		case rblx.AuthorizationDenied:
			fmt.Println("[Reaper] Cookie of the snipe account is invalid")
		case rblx.TooManyRequests:
			fmt.Println("[Reaper] Too many requests sent to the economy API")
		}

		fetchChannel <- nil
		return
	}

	fetchChannel <- resellerDetails
}

// BuyItem fetches the lowest priced item seller's userassetid and sellerid, then tries to buy that item for profit
func BuyItem(purchaseDetails *rblx.PurchasePost, snipeAccountSession *rblx.RBLXSession) {
	// Fetch the lowest reseller with a x thread race condition
	fetchChannel := make(chan *rblx.ResellersResponse)
	for i := 0; i < 3; i++ {
		go FetchResellers(purchaseDetails, snipeAccountSession, fetchChannel)
	}

	resellerDetails := <-fetchChannel

	//fmt.Println(resellerDetails)
	if resellerDetails == nil {
		return
	}

	if resellerDetails.Data[0].Price != purchaseDetails.ExpectedPrice {
		fmt.Printf("[Reaper] Price changed for asset %d while trying to snipe it (Old price: %d | New price: %d). \n", purchaseDetails.AssetID, purchaseDetails.ExpectedPrice, resellerDetails.Data[0].Price)
		return
	}
	//returns ua, sid
	purchaseDetails.UserAssetID = resellerDetails.Data[0].UserAssetID
	purchaseDetails.ExpectedSellerID = resellerDetails.Data[0].Seller.ID

	// Purchase the item
	bbc, _ := snipeAccountSession.PurchaseItem(purchaseDetails.AssetID, *purchaseDetails)

	//fetch assetid, sniped for, everything required for webhook
	assetidstring := strconv.FormatInt(purchaseDetails.AssetID, 10)
	getitemname, err := http.NewRequest("", fmt.Sprintf("https://api.roblox.com/marketplace/productinfo?assetId=%s", assetidstring), nil)
	if err != nil {
		fmt.Println("[REAPER] Failed to get item info to send webhook")
	}
	getitemname.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", globals.Config.Cookie))
	getitemnameres, err := client.Do(getitemname)
	if err != nil {
		fmt.Println("[REAPER] Failed to get item info to send webhook")
	}
	var itemname1 *GetItemInfo
	json.NewDecoder(getitemnameres.Body).Decode(&itemname1)
	itemname := itemname1.Name
	stringprice := strconv.FormatInt(purchaseDetails.ExpectedPrice, 10)

	if bbc.Purchased == true {
		//print info and send webhook
		fmt.Printf("[REAPER] Scythed %s for %s! \n", itemname, stringprice)
	} else {
		fmt.Printf("[REAPER] Failed to Scythe %s for %s! [REASON]: %s \n", itemname, stringprice, bbc.Reason)
	}



}
