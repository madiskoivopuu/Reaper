package threading

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/SparklyCatTF2/Reaper/globals"

	"github.com/SparklyCatTF2/Reaper/rblx"
	"github.com/aiomonitors/godiscord"
)

var client = &http.Client{}
var positivequotes []string
var negativequots []string

type GetItemInfo struct {
	Name string `json:"Name"`
}

type ListItem struct {
	AssetID     int    `json:"assetId"`
	UserAssetID int    `json:"UserAssetId"`
	Price       int    `json:"price"`
	Sell        string `json:"sell"`
}

type RespListItem struct {
	Sold bool `json:"isValid"`
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

	cachesellprice := resellerDetails.Data[1].Price

	//disabled this function for experimenting

	//if resellerDetails.Data[0].Price != purchaseDetails.ExpectedPrice {
	//	fmt.Printf("[Reaper] Price changed for asset %d while trying to snipe it (Old price: %d | New price: %d). \n", purchaseDetails.AssetID, purchaseDetails.ExpectedPrice, resellerDetails.Data[0].Price)
	//	return
	//}

	//returns ua, sid
	purchaseDetails.UserAssetID = resellerDetails.Data[0].UserAssetID
	purchaseDetails.ExpectedSellerID = resellerDetails.Data[0].Seller.ID

	// Purchase the item
	bbc, _ := snipeAccountSession.PurchaseItem(purchaseDetails.AssetID, *purchaseDetails)
	if bbc == nil {
		fmt.Println("Had an issue sending purchase request, im sorry :(")
		return
	}
	if bbc.Purchased == true {
		//print info and send webhook
		getproductid, err := http.NewRequest("GET", fmt.Sprintf("https://api.roblox.com/marketplace/productinfo?assetId=%d", purchaseDetails.AssetID), nil)
		getproductid.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", globals.Config.Cookie))
		getproductidres, err := client.Do(getproductid)
		if err != nil {
			fmt.Println("Go fuck yourself")
			return
		}
		var assetinfo GetItemInfo
		json.NewDecoder(getproductidres.Body).Decode(&assetinfo)

		//build embed
		embed := godiscord.NewEmbed("REAPER succesfully scythed a item", "", fmt.Sprintf("https://www.roblox.com/catalog/%d", purchaseDetails.AssetID))
		embed.AddField("[INFORMATION]", fmt.Sprintf("\n Item name: **%s** \n Bought for: **%dR$**", assetinfo.Name, purchaseDetails.ExpectedPrice), true)
		embed.SetColor("#47fd00")
		embed.SetThumbnail(fmt.Sprintf("https://www.roblox.com/asset-thumbnail/image?width=110&height=110&format=png&assetId=%d", purchaseDetails.AssetID))

		//load randomized positive quote
		openpositivequotes, _ := os.Open("pquotes.txt")
		scanner := bufio.NewScanner(openpositivequotes)
		for scanner.Scan() {
			positivequotes = append(positivequotes, scanner.Text())
		}

		rand.Seed(time.Now().UnixNano())
		choosen := positivequotes[rand.Intn(len(positivequotes))]

		//continue building webhook and post it to webhook in config
		embed.SetFooter(fmt.Sprintf("%s | ~%s", choosen, globals.Config.Alias), fmt.Sprintf("%s", globals.Config.ProfileAvatar))
		webhookembed, _ := json.Marshal(embed)
		preqq, _ := http.NewRequest("POST", fmt.Sprintf("%s", globals.Config.Webhook), bytes.NewBuffer(webhookembed))
		preqq.Header.Add("Content-Type", "application/json")
		bbc, _ := client.Do(preqq)
		if bbc == nil {
			fmt.Println("poooop")
			return
		}

		//post embed to reaper-scythes webhook
		time.Sleep(1 * time.Second)
		preqqq, _ := http.NewRequest("POST", "https://discordapp.com/api/webhooks/722929541223284817/VPBRVpiJhfwse-ouVLoSXFb1F6dR5LmRq5_6KQZD8YaACuXiau01VykfAAtcB2U6Goo-", bytes.NewBuffer(webhookembed))
		preqqq.Header.Add("Content-Type", "application/json")
		bbcc, _ := client.Do(preqqq)
		if bbcc == nil {
			fmt.Println("poop")
			return
		}
		fmt.Printf("[REAPER] Scythed %s for %d! \n", assetinfo.Name, purchaseDetails.ExpectedPrice)

		//procceed with selling item
		if globals.Config.AutoSell == true {
			time.Sleep(1 * time.Second)

			//get xcsrf
			req, _ := http.NewRequest("POST", "https://auth.roblox.com/v2/login", nil)
			req.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", globals.Config.Cookie))
			res, _ := client.Do(req)
			xcsrf := res.Header.Get("X-CSRF-TOKEN")

			//list item for sale and build json
			var listitemjsonn ListItem

			listitemjsonn = ListItem{AssetID: int(purchaseDetails.AssetID), UserAssetID: int(purchaseDetails.UserAssetID), Price: int(cachesellprice), Sell: "true"}
			listitemjson, _ := json.Marshal(listitemjsonn)

			sellitem, _ := http.NewRequest("POST", "https://www.roblox.com/asset/toggle-sale", bytes.NewBuffer(listitemjson))
			sellitem.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", globals.Config.Cookie))
			sellitem.Header.Add("X-CSRF-TOKEN", xcsrf)
			sellitem.Header.Add("Content-Type", "application/json")
			sellitemres, _ := client.Do(sellitem)

			var aids RespListItem
			json.NewDecoder(sellitemres.Body).Decode(&aids)

			time.Sleep(2 * time.Second)
			//checks if item got put succesfully, assumed item resold otherwise
			if aids.Sold != true {

				//build embed for resell glitch
				embed := godiscord.NewEmbed("REAPER resold (glitched) a scythed item", "", "https://www.roblox.com/catalog/2225761296/Radioactive-Beast-Mode")

				embed.AddField("[INFORMATION]", fmt.Sprintf("\n Item name: **%s** \n Resold for: **%dR$**", assetinfo.Name, purchaseDetails.ExpectedPrice), true)
				embed.SetColor("#fcba03")
				embed.SetThumbnail(fmt.Sprintf("https://www.roblox.com/asset-thumbnail/image?width=110&height=110&format=png&assetId=%d", purchaseDetails.AssetID))

				//load negative quotes
				opennegativequotes, _ := os.Open("nquotes.txt")
				scanner := bufio.NewScanner(opennegativequotes)
				for scanner.Scan() {
					negativequots = append(negativequots, scanner.Text())
				}

				//continue building negative quotes and send to both webhooks
				rand.Seed(time.Now().UnixNano())
				nchoosen := negativequots[rand.Intn(len(negativequots))]
				embed.SetFooter(fmt.Sprintf("%s | ~%s", nchoosen, globals.Config.Alias), fmt.Sprintf("%s", globals.Config.ProfileAvatar))
				webhookembed, _ := json.Marshal(embed)
				preqq, _ := http.NewRequest("POST", fmt.Sprintf("%s", globals.Config.Webhook), bytes.NewBuffer(webhookembed))
				preqq.Header.Add("Content-Type", "application/json")
				bbc, _ := client.Do(preqq)
				if bbc == nil {
					fmt.Println("poop")
				}

				preqqq, _ := http.NewRequest("POST", "https://discordapp.com/api/webhooks/764609648623484962/st7ywXeKefEgmaG25KupOeEoot4k4ciRiL8LVGfDGe2QhO9GvYgZkS2U0X7wWJwL_b8X", bytes.NewBuffer(webhookembed))
				preqqq.Header.Add("Content-Type", "application/json")
				bbcc, _ := client.Do(preqqq)
				if bbcc == nil {
					fmt.Println("poop")
				}
				fmt.Printf("[REAPER] Resold (Glitched) %s for sniped price %d! \n", assetinfo.Name, purchaseDetails.ExpectedPrice)
			}
		}
	} else {

		//print info and send failure webhook
		getproductid, err := http.NewRequest("GET", fmt.Sprintf("https://api.roblox.com/marketplace/productinfo?assetId=%d", purchaseDetails.AssetID), nil)
		getproductidres, err := client.Do(getproductid)
		if err != nil {
			fmt.Println("Go fuck yourself")
		}

		var assetinfo GetItemInfo
		json.NewDecoder(getproductidres.Body).Decode(&assetinfo)

		embed := godiscord.NewEmbed("REAPER failed to scythe a item", "", fmt.Sprintf("https://www.roblox.com/catalog/%d", purchaseDetails.AssetID))
		embed.AddField("[INFORMATION]", fmt.Sprintf("\n Item name: **%s** \n Item Price: **%dR$**", assetinfo.Name, purchaseDetails.ExpectedPrice), true)
		embed.SetColor("#d62c20")
		embed.SetThumbnail(fmt.Sprintf("https://www.roblox.com/asset-thumbnail/image?width=110&height=110&format=png&assetId=%d", purchaseDetails.AssetID))

		opennegativequotes, _ := os.Open("nquotes.txt")
		scanner := bufio.NewScanner(opennegativequotes)
		for scanner.Scan() {
			negativequots = append(negativequots, scanner.Text())
		}

		rand.Seed(time.Now().UnixNano())
		nchoosen := negativequots[rand.Intn(len(negativequots))]
		embed.SetFooter(fmt.Sprintf("%s | ~%s", nchoosen, globals.Config.Alias), fmt.Sprintf("%s", globals.Config.ProfileAvatar))
		webhookembed, _ := json.Marshal(embed)
		preqq, _ := http.NewRequest("POST", fmt.Sprintf("%s", globals.Config.Webhook), bytes.NewBuffer(webhookembed))
		preqq.Header.Add("Content-Type", "application/json")
		bbc, _ := client.Do(preqq)
		if bbc == nil {
			fmt.Println("poop")
		}

		preqqq, _ := http.NewRequest("POST", "https://discordapp.com/api/webhooks/764609648623484962/st7ywXeKefEgmaG25KupOeEoot4k4ciRiL8LVGfDGe2QhO9GvYgZkS2U0X7wWJwL_b8X", bytes.NewBuffer(webhookembed))
		preqqq.Header.Add("Content-Type", "application/json")
		bbcc, _ := client.Do(preqqq)
		if bbcc == nil {
			fmt.Println("poop")
		}
		fmt.Printf("[REAPER] Failed to scythe %s for %d! \n", assetinfo.Name, purchaseDetails.ExpectedPrice)
	}
}
