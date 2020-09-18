package globals

type config struct {
	Cookie           string  `json:"cookie"`
	IDsPerThread     int64   `json:"idsPerThread"`
	ThreadMultiplier int64   `json:"threadMultiplier"`
	Webhook          string  `json:"webhook"`
	WebhookMention   string  `json:"webhookMention"`
	ProfitPercent    float64 `json:"profitPercent"`
	AutoSell         bool    `json:"autoSell"`
}

var (
	Config            config
	CachedTokens      = make(map[string]string, 0)
	CachedPrices      = make(map[int64]int64, 0)
	CachedProductIDs  = make(map[int64]int64, 0)
	PriceCheckCookies []string
	Token             map[string]string
)
