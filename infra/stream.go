package infra

type Feed chan []byte

type PayloadTrade struct {
	Symbol string

	Price               float64
	Quantity            float64
	UNIXTimeMiliseconds int64
}

type ChFeed chan PayloadTrade

type StreamData struct {
	Stream string
	Feed   ChFeed
}

type Streams struct {
	Symbols []string
	Feed    Feed
}
