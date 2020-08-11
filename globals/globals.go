package globals

type config struct {
	Cookies        []string `json:"cookies"`
	Webhook        string   `json:"webhook"`
	WebhookMention string   `json:"webhookMention"`
	ProfitPercent  float64  `json:"profitPercent"`
	AutoSell       bool     `json:"autoSell"`
}

var (
	Config       config
	CachedTokens = make(map[string]string, 0)
)
