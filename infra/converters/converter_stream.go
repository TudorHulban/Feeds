package converters

import (
	"bnb/infra"

	"log"

	"github.com/tidwall/gjson"
)

type ConvertorStreams struct {
	processors []infra.IProcessor

	payload chan []byte
	stop    chan struct{}
}

var _ infra.IConverter = &ConvertorStreams{}

func NewStreamsConverter(procs ...infra.IProcessor) *ConvertorStreams {
	return &ConvertorStreams{
		processors: procs,

		payload: make(chan []byte),
		stop:    make(chan struct{}),
	}
}

// Convert Method converts Binance messages and pushes them further to a processor.
func (t *ConvertorStreams) Convert(locationOffsetMiliseconds int64) {
	for _, proc := range t.processors {
		go proc.Listen(locationOffsetMiliseconds)
		defer proc.Terminate()
	}

loop:
	for {
		select {
		case <-t.stop:
			{
				log.Println("stopping converter")
				break loop
			}

		case streamPayload := <-t.payload:
			{
				if len(streamPayload) == 0 {
					continue
				}

				result := gjson.GetManyBytes(streamPayload, "stream", "data.s", "data.T", "data.q", "data.p")

				for _, proc := range t.processors {
					if proc.Payload().Stream == result[0].String() {
						proc.Payload().Feed <- infra.PayloadTrade{
							Symbol:              result[1].String(),
							UNIXTimeMiliseconds: result[2].Int(),
							Price:               result[3].Float(),
							Quantity:            result[4].Float(),
						}
					}
				}
			}
		}
	}
}

func (t *ConvertorStreams) Payload() infra.Streams {
	symbols := make([]string, len(t.processors), len(t.processors))

	for ix, symbol := range t.processors {
		symbols[ix] = symbol.Payload().Stream
	}

	return infra.Streams{
		Symbols: symbols,
		Feed:    t.payload,
	}
}

func (t *ConvertorStreams) Terminate() {
	defer t.cleanUp()

	t.stop <- struct{}{}
}

func (t *ConvertorStreams) cleanUp() {
	close(t.stop)
}
