package threading

import (
	"fmt"

	"github.com/SparklyCatTF2/Reaper/globals"
	"github.com/SparklyCatTF2/Reaper/rblx"
)

// FetchResellers gets resellers for an asset ID and then passes the details through a channel
func FetchResellers(purchaseDetails *rblx.PurchasePost, snipeAccountSession *rblx.RBLXSession, fetchChannel chan *rblx.ResellersResponse) {
	s1 := globals.GetTimeInMs()
	resellerDetails, err := snipeAccountSession.GetResellers(purchaseDetails.AssetID)
	s2 := globals.GetTimeInMs() - s1
	fmt.Println(s2)

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
	// Fetch the lowest reseller with a 3 thread race condition
	fetchChannel := make(chan *rblx.ResellersResponse)
	for i := 0; i < 3; i++ {
		go FetchResellers(purchaseDetails, snipeAccountSession, fetchChannel)
	}

	resellerDetails := <-fetchChannel

	//fmt.Println(resellerDetails)
	if resellerDetails == nil { return }

	if resellerDetails.Data[0].Price != purchaseDetails.ExpectedPrice {
		fmt.Printf("[Reaper] Price changed for asset %d while trying to snipe it (Old price: %d | New price: %d).", purchaseDetails.AssetID, purchaseDetails.ExpectedPrice, resellerDetails.Data[0].Price)
		return
	}

	// Purchase the item

}