package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/SparklyCatTF2/Reaper/globals"
	"github.com/SparklyCatTF2/Reaper/threading"
)

var (
	priceCheckCookies []string
	assetIds          []int64
)

func main() {
	//blockedUserAssetIds := make(map[int64]int64, 0)

	// Load Config.cfg
	configFile, configFileError := ioutil.ReadFile("./settings/config.json")
	if configFileError != nil {
		fmt.Printf("[Reaper] Failed to load config.json - %s", configFileError.Error())
		return
	}
	jsonParseError := json.Unmarshal(configFile, &globals.Config)
	if jsonParseError != nil {
		fmt.Printf("[Reaper] Failed to parse config.json - %s", configFileError.Error())
		return
	}

	// Load all price check cookies
	cookiesFile, cookiesFileError := ioutil.ReadFile("./settings/cookies.txt")
	if cookiesFileError != nil {
		fmt.Printf("[Reaper] Failed to load cookies.txt - %s", cookiesFileError.Error())
		return
	}
	globals.PriceCheckCookies = strings.Split(string(cookiesFile), "\n")

	// Load asset IDs
	assetIdsFile, assetsFileError := ioutil.ReadFile("./settings/asset_ids.txt")
	if assetsFileError != nil {
		fmt.Printf("[Reaper] Failed to load asset IDs - %s", assetsFileError.Error())
		return
	}
	for line, idStr := range bytes.Split(assetIdsFile, []byte{'\n'}) {
		id, parseError := strconv.ParseInt(string(idStr), 10, 64)
		if parseError != nil {
			fmt.Printf("[Reaper] Failed to convert asset id on line %d - %s", line, parseError.Error())
			return
		}

		assetIds = append(assetIds, id)
	}

	// Passing asset IDs to the snipe threads
	currentAssetNum := int64(1)
	tempAssetIds := []int64{}
	// snipeChannel := make(chan PurchaseStruct)
	for _, assetId := range assetIds {
		if currentAssetNum < globals.Config.IDsPerThread {
			tempAssetIds = append(tempAssetIds, assetId)
			currentAssetNum++
		} else {
			// add the remaining asset ids to the list
			tempAssetIds = append(tempAssetIds, assetId)

			for m := int64(0); m < globals.Config.ThreadMultiplier; m++ {
				go threading.SnipeThread(tempAssetIds /*, snipeChannel */)
			}

			tempAssetIds = nil
		}
	}

	/*for {
		snipe := <-snipeChannel
		// create a separate thread for purchasing
		// add userassetid to blocked list
	}*/
}
