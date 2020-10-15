package threading

import (
	"net/http"
	"time"

	"github.com/SparklyCatTF2/Reaper/rblx"
)

// ConnectionThread keeps the connection to economy.roblox.com open on a *rblx.RBLXSession instance
func ConnectionThread(snipeAccountSession *rblx.RBLXSession) {
	for {
		req, _ := http.NewRequest("GET", "https://economy.roblox.com/v1/assets/64082730/resellers?limit=10", nil)
		req.Header.Add("Cookie", ".ROBLOSECURITY="+snipeAccountSession.Cookie)
		resp, respErr := snipeAccountSession.Client.Do(req)
		if respErr != nil { continue }
		resp.Body.Close()

		req2, _ := http.NewRequest("POST", "https://economy.roblox.com/v1/purchases/products/256946335", nil)
		req2.Header.Add("Cookie", ".ROBLOSECURITY="+snipeAccountSession.Cookie)
		resp2, respErr := snipeAccountSession.Client.Do(req2)
		if respErr != nil { continue }
		resp2.Body.Close()


		time.Sleep(6 * time.Second)
	}
}