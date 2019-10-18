package drivers

import (
	"bytes"
	"encoding/binary"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

const (
	Adxl345AccAddress       = 0x53
	Adxl345DataAddress      = 0x32
	Adxl345ConfModeAddress  = 0x2D
	Adxl345ConfModeValue    = 0x08 // measurement mode
	Adxl345ConfRangeAddress = 0x31
	Adxl345ConfRangeValue   = 0x00 // +/- 2g
)

type ADXL345Driver struct {
	name          string
	connector     i2c.Connector
	connection    i2c.Connection
	Accelerometer i2c.ThreeDData
	i2c.Config
	gobot.Eventer
}

func NewADXL345Driver(a i2c.Connector, options ...func(i2c.Config)) *ADXL345Driver {
	m := &ADXL345Driver{
		name:      gobot.DefaultName("ADXL345"),
		connector: a,
		Config:    i2c.NewConfig(),
		Eventer:   gobot.NewEventer(),
	}

	for _, option := range options {
		option(m)
	}

	return m
}

func (d *ADXL345Driver) Name() string {
	return d.name
}

func (d *ADXL345Driver) SetName(s string) {
	d.name = s
}

func (d *ADXL345Driver) Start() error {
	if err := d.initialize(); err != nil {
		return err
	}

	return nil
}

func (d *ADXL345Driver) Halt() error {
	return nil
}

func (d *ADXL345Driver) Connection() gobot.Connection {
	return d.connection.(gobot.Connection)
}

func (d *ADXL345Driver) initialize() (err error) {
	bus := d.GetBusOrDefault(d.connector.GetDefaultBus())
	address := d.GetAddressOrDefault(Adxl345AccAddress)

	d.connection, err = d.connector.GetConnection(address, bus)
	if err != nil {
		return err
	}

	// setMode
	if _, err = d.connection.Write([]byte{Adxl345ConfModeAddress, Adxl345ConfModeValue}); err != nil {
		return
	}

	// setFullScaleAccelRange
	if _, err = d.connection.Write([]byte{Adxl345ConfRangeAddress, Adxl345ConfRangeValue}); err != nil {
		return
	}

	return nil
}

// GetData fetches the latest data from the ADXL345
func (h *ADXL345Driver) GetData() (err error) {
	if _, err = h.connection.Write([]byte{Adxl345DataAddress}); err != nil {
		return
	}

	data := make([]byte, 6)
	_, err = h.connection.Read(data)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	return binary.Read(buf, binary.BigEndian, &h.Accelerometer)
}
