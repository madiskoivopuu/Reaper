package threading

import (
	"net/http"

	"github.com/SparklyCatTF2/Reaper/rblx"
)

func GrabToken(session *rblx.RBLXSession, loopForever bool) {
	httpClient := http.Client{}

	// do while loop
	for continueLoop := true; continueLoop; continueLoop = loopForever {
		req, _ := http.NewRequest("POST", "https://catalog.roblox.com/v1/catalog/items/details", nil)
		req.Header.Add("Cookie", ".ROBLOSECURITY="+session.Cookie)
		req.Header.Add("Content-Type", "application/json")

		resp, respError := httpClient.Do(req)
		if respError != nil {
			continue
		}

		if token := resp.Header.Get("X-CSRF-TOKEN"); token != "" {
			session.XCSRFToken = token
		}
	}
}
