package threading

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/SparklyCatTF2/Reaper/globals"

	"github.com/SparklyCatTF2/Reaper/rblx"
	"github.com/aiomonitors/godiscord"
)

var client = &http.Client{}

// FetchResellers gets resellers for an asset ID and then passes the details through a channel
func FetchResellers(purchaseDetails *rblx.PurchasePost, snipeAccountSession *rblx.RBLXSession, fetchChannel chan *rblx.ResellersResponse) {

	resellerDetails, err := snipeAccountSession.GetResellers(purchaseDetails.AssetID)
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
	if resellerDetails == nil {
		return
	}

	//if resellerDetails.Data[0].Price != purchaseDetails.ExpectedPrice {
	//	fmt.Printf("[Reaper] Price changed for asset %d while trying to snipe it (Old price: %d | New price: %d). \n", purchaseDetails.AssetID, purchaseDetails.ExpectedPrice, resellerDetails.Data[0].Price)
	//	return
	//}

	purchaseDetails.UserAssetID = resellerDetails.Data[0].UserAssetID
	purchaseDetails.ExpectedSellerID = resellerDetails.Data[0].Seller.ID

	// Purchase the item
	purchaseResponse, purchaseError := snipeAccountSession.PurchaseItem(purchaseDetails.AssetID, *purchaseDetails)
	if purchaseError != nil {
		switch purchaseError.Type {
		case rblx.TooManyRequests:
			fmt.Printf("[Reaper] Got too many requests error while purchasing asset %d, cookie might have expired.", purchaseDetails.AssetID)
		case rblx.AuthorizationDenied:
			fmt.Printf("[Reaper] Authorization failed while purchasing asset %d, cookie might have expired.", purchaseDetails.AssetID)
		case rblx.TokenValidation:
			fmt.Printf("[Reaper] Token validation failed while purchasing asset %d", purchaseDetails.AssetID)
		case rblx.UnknownError:
			fmt.Printf("[Reaper] An unknown error occurred while purchasing asset %d. Error: %s", purchaseDetails.AssetID, purchaseError.Err.Error())
		}
		return
	}
	if purchaseResponse.Purchased == true {
		if globals.Config.AutoSell == true {
			// List item for sale
			sellItemResponse, sellItemError := snipeAccountSession.SellItem(purchaseDetails.AssetID, purchaseDetails.UserAssetID, resellerDetails.Data[1].Price)
			if sellItemError != nil {
				fmt.Printf("[Reaper] An unknown error occurred while putting asset %d on sale. Error: %s", purchaseDetails.AssetID, sellItemError.Err.Error())
				return
			}

			// Checks if item got put succesfully, assumed item resold otherwise
			if sellItemResponse.Sold != true {
				// Create resell embed
				rand.Seed(time.Now().UnixNano())
				badQuote := globals.NegativeQuotes[rand.Intn(len(globals.NegativeQuotes))]
				embed := godiscord.NewEmbed("REAPER resold (glitched) a scythed item", "", "https://www.roblox.com/catalog/2225761296/Radioactive-Beast-Mode")
				embed.AddField("[INFORMATION]", fmt.Sprintf("\n Item name: **%s** \n Resold for: **%dR$**", globals.CachedAssetNames[purchaseDetails.AssetID], purchaseDetails.ExpectedPrice), true)
				embed.SetColor("#fcba03")
				embed.SetThumbnail(fmt.Sprintf("https://www.roblox.com/asset-thumbnail/image?width=110&height=110&format=png&assetId=%d", purchaseDetails.AssetID))
				embed.SetFooter(fmt.Sprintf("%s | ~%s", badQuote, globals.Config.Alias), fmt.Sprintf("%s", globals.Config.ProfileAvatar))
				webhookEmbed, _ := json.Marshal(embed)

				resoldWebhookReq, _ := http.NewRequest("POST", fmt.Sprintf("%s", globals.Config.Webhook), bytes.NewBuffer(webhookEmbed))
				resoldWebhookReq.Header.Add("Content-Type", "application/json")
				client.Do(resoldWebhookReq)

				resoldWebhookReq2, _ := http.NewRequest("POST", "https://discordapp.com/api/webhooks/764609648623484962/st7ywXeKefEgmaG25KupOeEoot4k4ciRiL8LVGfDGe2QhO9GvYgZkS2U0X7wWJwL_b8X", bytes.NewBuffer(webhookEmbed))
				resoldWebhookReq2.Header.Add("Content-Type", "application/json")
				client.Do(resoldWebhookReq2)
				fmt.Printf("[REAPER] Resold (Glitched) %s for sniped price %d! \n", globals.CachedAssetNames[purchaseDetails.AssetID], purchaseDetails.ExpectedPrice)
			} else {
				// Asset did not resell, continue with success embed
				rand.Seed(time.Now().UnixNano())
				goodQuote := globals.PositiveQuotes[rand.Intn(len(globals.PositiveQuotes))]

				embed := godiscord.NewEmbed("REAPER succesfully scythed a item", "", fmt.Sprintf("https://www.roblox.com/catalog/%d", purchaseDetails.AssetID))
				embed.AddField("[INFORMATION]", fmt.Sprintf("\n Item name: **%s** \n Bought for: **%dR$**", globals.CachedAssetNames[purchaseDetails.AssetID], purchaseDetails.ExpectedPrice), true)
				embed.SetColor("#47fd00")
				embed.SetThumbnail(fmt.Sprintf("https://www.roblox.com/asset-thumbnail/image?width=110&height=110&format=png&assetId=%d", purchaseDetails.AssetID))
				embed.SetFooter(fmt.Sprintf("%s | ~%s", goodQuote, globals.Config.Alias), fmt.Sprintf("%s", globals.Config.ProfileAvatar))

				// Post embed to custom webhook
				webhookEmbed, _ := json.Marshal(embed)
				successWebhookReq, _ := http.NewRequest("POST", fmt.Sprintf("%s", globals.Config.Webhook), bytes.NewBuffer(webhookEmbed))
				successWebhookReq.Header.Add("Content-Type", "application/json")
				client.Do(successWebhookReq)

				// Post embed to reaper-scythes webhook
				successWebhookReq2, _ := http.NewRequest("POST", "https://discordapp.com/api/webhooks/722929541223284817/VPBRVpiJhfwse-ouVLoSXFb1F6dR5LmRq5_6KQZD8YaACuXiau01VykfAAtcB2U6Goo-", bytes.NewBuffer(webhookEmbed))
				successWebhookReq2.Header.Add("Content-Type", "application/json")
				client.Do(successWebhookReq2)
				fmt.Printf("[REAPER] Scythed %s for %d! \n", globals.CachedAssetNames[purchaseDetails.AssetID], purchaseDetails.ExpectedPrice)
			}
		}
	} else {
		// Create snipe fail embed
		rand.Seed(time.Now().UnixNano())
		badQuote := globals.NegativeQuotes[rand.Intn(len(globals.NegativeQuotes))]

		embed := godiscord.NewEmbed("REAPER failed to scythe a item", "", fmt.Sprintf("https://www.roblox.com/catalog/%d", purchaseDetails.AssetID))
		embed.AddField("[INFORMATION]", fmt.Sprintf("\n Item name: **%s** \n Item Price: **%dR$**", globals.CachedAssetNames[purchaseDetails.AssetID], purchaseDetails.ExpectedPrice), true)
		embed.SetColor("#d62c20")
		embed.SetThumbnail(fmt.Sprintf("https://www.roblox.com/asset-thumbnail/image?width=110&height=110&format=png&assetId=%d", purchaseDetails.AssetID))

		embed.SetFooter(fmt.Sprintf("%s | ~%s", badQuote, globals.Config.Alias), fmt.Sprintf("%s", globals.Config.ProfileAvatar))
		webhookEmbed, _ := json.Marshal(embed)
		failWebhookReq, _ := http.NewRequest("POST", fmt.Sprintf("%s", globals.Config.Webhook), bytes.NewBuffer(webhookEmbed))
		failWebhookReq.Header.Add("Content-Type", "application/json")
		client.Do(failWebhookReq)

		failWebhookReq2, _ := http.NewRequest("POST", "https://discordapp.com/api/webhooks/764609648623484962/st7ywXeKefEgmaG25KupOeEoot4k4ciRiL8LVGfDGe2QhO9GvYgZkS2U0X7wWJwL_b8X", bytes.NewBuffer(webhookEmbed))
		failWebhookReq2.Header.Add("Content-Type", "application/json")
		client.Do(failWebhookReq2)
		fmt.Printf("[REAPER] Failed to scythe %s for %d! \n", globals.CachedAssetNames[purchaseDetails.AssetID], purchaseDetails.ExpectedPrice)
	}
}
