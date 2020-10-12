package threading

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/SparklyCatTF2/Reaper/globals"
	"github.com/SparklyCatTF2/Reaper/rblx"
)
var counter = 0
func SnipeThread(assetIDs []int64, snipeChannel chan *rblx.PurchasePost) {
	emptystr := ""
	httpclient := &http.Client{}
	cachedPrices := make(map[int64]int64, 0)
	rand.Seed(time.Now().UnixNano())

	// Cache the token for the current and next price check cookie
	currentRobloxSession := &rblx.RBLXSession{Cookie: globals.PriceCheckCookies[rand.Intn(len(globals.PriceCheckCookies))], Client: httpclient, XCSRFToken: &emptystr}
	nextRobloxSession := &rblx.RBLXSession{Cookie: globals.PriceCheckCookies[rand.Intn(len(globals.PriceCheckCookies))], Client: httpclient, XCSRFToken: &emptystr}
	rblxsession := currentRobloxSession
	GrabToken(rblxsession, false)
	go GrabToken(nextRobloxSession, false)

	for {
		detailsResponse, err := rblxsession.GetCatalogDetails(assetIDs)
		if err != nil {
			//fmt.Println(err)
			// Rate limit, change price check cookie
			switch err.Type {
			case rblx.TooManyRequests:
				currentRobloxSession = nextRobloxSession
				nextRobloxSession = &rblx.RBLXSession{Cookie: globals.PriceCheckCookies[rand.Intn(len(globals.PriceCheckCookies))], Client: httpclient, XCSRFToken: &emptystr}
				go GrabToken(nextRobloxSession, false)
			case rblx.AuthorizationDenied:
				fmt.Printf("[Reaper] Invalid price check cookie %s", rblxsession.Cookie)
			case rblx.TokenValidation:
				GrabToken(rblxsession, false)
			}
			continue
		}

		// Loop over the items & send the purchase details to main thread if snipe is profitable
		for _, item := range detailsResponse.Data {
			//counter += 1
			//fmt.Println(counter)
			fmt.Printf("LowestPrice: %d | Cached Price: %d \n", item.LowestPrice, cachedPrices[item.ID])
			if item.LowestPrice == 0 {
				fmt.Println("Lowest Price was 0!")
				return
			}
			if cachedPrices[item.ID] == 0 {
				fmt.Println("CachedPrice was 0!")
				return
			}
			if item.LowestPrice < cachedPrices[item.ID] {
				getpercent := float64((30 * item.LowestPrice) / 100)
				oldPriceAfterTax := float64(cachedPrices[item.ID])
				oldPriceAfterTax -= getpercent
				profitMargin := oldPriceAfterTax - float64(item.LowestPrice)
				profitPercent := profitMargin / float64(item.LowestPrice)
				if profitPercent >= globals.Config.ProfitPercent {
					purchaseStruct := &rblx.PurchasePost{AssetID: item.ID, ExpectedCurrency: 1, ExpectedPrice: item.LowestPrice}
					snipeChannel <- purchaseStruct
				} else {
					cachedPrices[item.ID] = item.LowestPrice
				}
			} else {
				cachedPrices[item.ID] = item.LowestPrice
			}
		}

	}
}
