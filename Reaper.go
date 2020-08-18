package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/SparklyCatTF2/Vulkan/globals"
)

var (
	priceCheckCookies []string
	assetIds          []int64
)

func main() {
	// Load Config.cfg
	configFile, configFileError := ioutil.ReadFile("./settings/config.json")
	if configFileError != nil {
		fmt.Printf("[Vulkan] Failed to load config.json - %s", configFileError.Error())
		return
	}
	jsonParseError := json.Unmarshal(configFile, &globals.Config)
	if jsonParseError != nil {
		fmt.Printf("[Vulkan] Failed to parse config.json - %s", configFileError.Error())
		return
	}

	// Load all price check cookies
	cookiesFile, cookiesFileError := ioutil.ReadFile("./settings/cookies.txt")
	if cookiesFileError != nil {
		fmt.Printf("[Vulkan] Failed to load cookies.txt - %s", cookiesFileError.Error())
		return
	}
	priceCheckCookies = strings.Split(string(cookiesFile), "\n")

	// Load asset IDs
	assetIdsFile, assetsFileError := ioutil.ReadFile("./settings/asset_ids.txt")
	if assetsFileError != nil {
		fmt.Printf("[Vulkan] Failed to load asset IDs - %s", assetsFileError.Error())
		return
	}
	for line, idStr := range bytes.Split(assetIdsFile, []byte{'\n'}) {
		id, parseError := strconv.ParseInt(string(idStr), 10, 64)
		if parseError != nil {
			fmt.Printf("[Vulkan] Failed to convert asset id on line %d - %s", line, parseError.Error())
			return
		}

		assetIds = append(assetIds, id)
	}

}
