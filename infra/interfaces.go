package infra

import "io"

type IConverter interface {
	Convert(locationOffsetMiliseconds int64)
	Payload() Streams // provides feed channel to receive exchange payload
	Terminate()
}

type IProcessor interface {
	Listen(locationOffsetMiliseconds int64)

	// provides symbol and channel to receive trade payload
	Payload() StreamData
	SendBufferTo(io.Writer)
	Terminate()
}
