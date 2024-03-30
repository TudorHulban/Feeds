package converters

import (
	"bnb/infra"
	"log"

	"github.com/tidwall/gjson"
)

type ConvertorBinanceTrade struct {
	processor infra.IProcessor

	payload chan []byte
	stop    chan struct{}
}

func NewConverterBinanceTrade(proc infra.IProcessor) *ConvertorBinanceTrade {
	return &ConvertorBinanceTrade{
		processor: proc,
		payload:   make(chan []byte),
		stop:      make(chan struct{}),
	}
}

// Convert Method converts Binance messages and pushes them further to a processor.
func (t *ConvertorBinanceTrade) Convert() {
	go t.processor.Listen(0)
	defer t.processor.Terminate()

	processorPayload := t.processor.Payload().Feed

loop:
	for {
		select {
		case <-t.stop:
			{
				log.Println("stopping converter")
				break loop
			}

		case payload := <-t.payload:
			{
				if len(payload) == 0 {
					continue
				}

				result := gjson.GetManyBytes(payload, "s", "T", "q", "p")

				processorPayload <- infra.PayloadTrade{
					Symbol:              result[0].String(),
					UNIXTimeMiliseconds: result[1].Int(),
					Price:               result[2].Float(),
					Quantity:            result[3].Float(),
				}
			}
		}
	}
}

func (t *ConvertorBinanceTrade) Payload() infra.Feed {
	return t.payload
}

func (t *ConvertorBinanceTrade) Terminate() {
	defer t.cleanUp()

	t.stop <- struct{}{}
}

func (t *ConvertorBinanceTrade) cleanUp() {
	close(t.stop)
}
