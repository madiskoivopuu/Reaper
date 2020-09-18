package threading

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/SparklyCatTF2/Reaper/globals"
	"github.com/SparklyCatTF2/Reaper/rblx"
)

func SnipeThread(assetIDs []int64) {
	httpclient := &http.Client{}
	rand.Seed(time.Now().UnixNano())
	rblxsession := &rblx.RBLXSession{}
	for {
		// Cache the token for the current and next price check cookie
		currentPriceCheckCookie := globals.PriceCheckCookies[rand.Intn(len(globals.PriceCheckCookies))]
		nextPriceCheckCookie := globals.PriceCheckCookies[rand.Intn(len(globals.PriceCheckCookies))]
		rblxsession = &rblx.RBLXSession{Cookie: currentPriceCheckCookie, Client: httpclient}
		go GrabToken(rblxsession, false)
		go GrabToken(&rblx.RBLXSession{Cookie: nextPriceCheckCookie, Client: httpclient}, false)

		detailsResponse, err := rblxsession.GetCatalogDetails(assetIDs)
		if err != nil {
			// Rate limit, change price check cookie
			switch err.Type {
			case rblx.TooManyRequests:
				currentPriceCheckCookie := nextPriceCheckCookie
				rblxsession.Cookie = currentPriceCheckCookie
				nextPriceCheckCookie := globals.PriceCheckCookies[rand.Intn(len(globals.PriceCheckCookies))]
				go GrabToken(&rblx.RBLXSession{Cookie: nextPriceCheckCookie, Client: httpclient}, false)
			case rblx.AuthorizationDenied:
				fmt.Printf("[Reaper] Invalid price check cookie %s", rblxsession.Cookie)
			}
		}

		for _, item := range detailsResponse.Data {
			if globals.CachedPrices[item.ID] <= item.LowestPrice {
				getpercent := float64((30 * item.LowestPrice) / 100)
				oldPriceAfterTax := float64(globals.CachedPrices[item.ID])
				oldPriceAfterTax -= getpercent
				profitMargin := oldPriceAfterTax - float64(item.LowestPrice)
				profitPercent := profitMargin / float64(item.LowestPrice)

				if profitPercent >= globals.Config.ProfitPercent {
					//purchaseStruct := &rblx.PurchasePost{ExpectedCurrency: 1, ExpectedPrice: item.LowestPrice}
					// send the details through channel
				} else {
					globals.CachedPrices[item.ID] = item.LowestPrice
				}
			}
		}

	}

	return
}
