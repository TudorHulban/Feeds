package infra

type PayloadTrade struct {
	Symbol string

	Price                float64
	Quantity             float64
	TimestampMiliseconds int64
}

type ChFeed chan PayloadTrade

type StreamData struct {
	Stream string
	Feed   ChFeed
}

type Streams struct {
	Symbols []string
	Feed    chan []byte
}
