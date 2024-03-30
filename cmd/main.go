package main

import (
	"bnb/domain/exchange"
	"bnb/infra/converters"
	"bnb/infra/processors"
	"fmt"
	"os"
	"strings"
)

func main() {
	processorBNB := processors.NewRolling("bnbusdt@trade", 1, os.Stdout)

	converterBNB := converters.NewStreamsConverter(processorBNB)

	urlStreams := rootStreamBinance + strings.Join(converterBNB.Payload().Symbols, "/")

	exch, errNew := exchange.NewExchangeGorilla(
		exchange.Config{
			URI: urlStreams,
		},
	)
	if errNew != nil {
		fmt.Println(errNew)

		os.Exit(1)
	}

	go exch.ReadMessages(converterBNB)
	exch.Work()

	converterBNB.Terminate()
}
