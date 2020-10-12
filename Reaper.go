package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/SparklyCatTF2/Reaper/globals"
	"github.com/SparklyCatTF2/Reaper/rblx"
	"github.com/SparklyCatTF2/Reaper/threading"
)

var (
	assetIds          []int64
	client = &http.Client{}
)

type GetProductId struct {
	ProductID   int64    `json:"ProductId"`
	Name        string   `json:"Name"`
}

func main() {
	// Load Config.cfg
	configFile, configFileError := ioutil.ReadFile("./settings/config.json")
	if configFileError != nil {
		fmt.Printf("[Reaper] Failed to load config.json - %s", configFileError.Error())
		return
	}
	jsonParseError := json.Unmarshal(configFile, &globals.Config)
	if jsonParseError != nil {
		fmt.Printf("[Reaper] Failed to parse config.json - %s", jsonParseError.Error())
		return
	}

	// Load all price check cookies
	opencookiesfile, _ := os.Open("./settings/cookies.txt")
	scanner := bufio.NewScanner(opencookiesfile)
	for scanner.Scan() {
		globals.PriceCheckCookies = append(globals.PriceCheckCookies, scanner.Text())
	}

	// Load asset IDs
	assetIdsFile, assetsFileError := os.Open("./settings/asset_ids.txt")
	if assetsFileError != nil {
		fmt.Printf("[Reaper] Failed to load asset IDs - %s", assetsFileError.Error())
		return
	}
	scanner1 := bufio.NewScanner(assetIdsFile)
	for scanner1.Scan() {
		jesus := []byte(scanner1.Text())
		for line, idStr := range bytes.Split(jesus, []byte{'\n'}) {
			id, parseError := strconv.ParseInt(string(idStr), 10, 64)
			if parseError != nil {
				fmt.Printf("[Reaper] Failed to convert asset id on line %d - %s", line, parseError.Error())
			}
			assetIds = append(assetIds, id)
		}
	}

	// Load positive & negative quotes
	filePositiveQuotes, _ := os.Open("pquotes.txt")
	scanner2 := bufio.NewScanner(filePositiveQuotes)
	for scanner2.Scan() {
		globals.PositiveQuotes = append(globals.PositiveQuotes, scanner2.Text())
	}

	fileNegativeQuotes, _ := os.Open("nquotes.txt")
	scanner3 := bufio.NewScanner(fileNegativeQuotes)
	for scanner3.Scan() {
		globals.NegativeQuotes = append(globals.NegativeQuotes, scanner3.Text())
	}

	// Cache ProductIDS of loaded assetIds
	for line, assetId := range assetIds {
		productIdReq, err := http.NewRequest("GET", fmt.Sprintf("https://api.roblox.com/marketplace/productinfo?assetId=%d", assetId), nil)
		if err != nil {
			fmt.Printf("[Reaper] Failed to create request to fetch ProductID for AssetID %s", line)
		}
		productIdReq.Header.Add("Cookie", fmt.Sprintf(".ROBLOSECURITY=%s", globals.Config.Cookie))
		productIdResp, productIdRespErr := client.Do(productIdReq)
		if productIdRespErr != nil {
			fmt.Printf("[Reaper] Failed to grab ProductID of AssetID of %s ", line)
		}
		var product *GetProductId
		json.NewDecoder(productIdResp.Body).Decode(&product)
		globals.CachedProductIDs[assetId] = product.ProductID
		globals.CachedAssetNames[assetId] = product.Name
	}


	// Start fetching token for snipe cookie & keeping connection to economy.roblox.com open
	snipeAccountSession := &rblx.RBLXSession{Cookie: globals.Config.Cookie, Client: &http.Client{}}
	go threading.GrabToken(snipeAccountSession, true)
	go threading.ConnectionThread(snipeAccountSession)

	// Passing asset IDs to the snipe threads
	currentAssetNum := int64(1)
	tempAssetIds := []int64{}
	snipeChannel := make(chan *rblx.PurchasePost)
	for _, assetID := range assetIds {
		if currentAssetNum < globals.Config.IDsPerThread {
			tempAssetIds = append(tempAssetIds, assetID)
			currentAssetNum++
		} else {
			// add the remaining asset ids to the list
			tempAssetIds = append(tempAssetIds, assetID)

			for m := int64(0); m < globals.Config.ThreadMultiplier; m++ {
				go threading.SnipeThread(tempAssetIds, snipeChannel)
				time.Sleep(1 * time.Millisecond)
			}

			tempAssetIds = nil
		}
	}

	for {
		purchaseDetails := <-snipeChannel
		// Check if sniping an asset is on cooldown
		if globals.BlockedAssetIds[purchaseDetails.AssetID] > globals.GetTimeInMs() {
			continue
		}
		globals.BlockedAssetIds[purchaseDetails.AssetID] = globals.GetTimeInMs() + 1*750 // add 1 second cooldown to assetid

		go threading.BuyItem(purchaseDetails, snipeAccountSession)
	}
}