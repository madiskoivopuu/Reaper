package threading

import (
	"net/http"
	"time"

	"github.com/SparklyCatTF2/Reaper/rblx"
)

// ConnectionThread keeps the connection to economy.roblox.com open on a *rblx.RBLXSession instance
func ConnectionThread(snipeAccountSession *rblx.RBLXSession) {
	for {
		req, _ := http.NewRequest("GET", "https://economy.roblox.com/v1/resale-tax-rate", nil)
		req.Header.Add("Cookie", ".ROBLOSECURITY="+snipeAccountSession.Cookie)
		resp, respErr := snipeAccountSession.Client.Do(req)
		if respErr != nil { continue }
		resp.Body.Close()

		time.Sleep(5 * time.Second)
	}
}