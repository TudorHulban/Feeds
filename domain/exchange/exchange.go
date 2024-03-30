package exchange

import (
	"bnb/infra"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type Config struct {
	RequestHeader       http.Header
	URI                 string
	SecondsPongInterval uint
}

type Exchange struct {
	connection *websocket.Conn

	url url.URL

	stop      chan struct{}
	interrupt chan os.Signal
}

func NewExchangeGorilla(cfg Config) (*Exchange, error) {
	url, errParse := url.Parse(cfg.URI)
	if errParse != nil {
		return nil, errParse
	}

	conn, _, errConn := websocket.DefaultDialer.Dial(url.String(), cfg.RequestHeader)
	if errConn != nil {
		return nil, errConn
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	return &Exchange{
			connection: conn,
			url:        *url,
			stop:       make(chan struct{}),
			interrupt:  interrupt,
		},
		nil
}

// ReadMessages Method reads websocket feed and pushes it to a converter payload channel.
func (e *Exchange) ReadMessages(conv infra.IConverter) {
	converterPayload := conv.Payload().Feed
	defer close(converterPayload)

	go conv.Convert(0)

loop:
	for {
		select {
		case <-e.interrupt:
			{
				fmt.Println("interrupt")
				break loop
			}

		default:
			{
				_, message, errRead := e.connection.ReadMessage()

				if errRead != nil {
					fmt.Println("read glitch:", errRead)

					return
				}

				converterPayload <- message
			}
		}
	}

	e.Terminate()
}

// Work Method blocking for work to be done.
func (e *Exchange) Work() {
	<-e.stop
}

func (e *Exchange) cleanUp() {
	close(e.interrupt)
	close(e.stop)
}

func (e *Exchange) Terminate() {
	defer e.cleanUp()

	e.stop <- struct{}{}
}
