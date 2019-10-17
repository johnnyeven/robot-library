package drivers

import (
	"gobot.io/x/gobot/drivers/aio"
	"gobot.io/x/gobot/drivers/gpio"
)

type Connection interface {
	gpio.DigitalWriter
	gpio.DigitalReader
	gpio.PwmWriter
	aio.AnalogReader
}
