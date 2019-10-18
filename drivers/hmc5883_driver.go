package drivers

import (
	"bytes"
	"encoding/binary"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

const (
	HMC5883CompassAddress  = 0x1E
	HMC5883DataAddress     = 0x03
	HMC5883ConfModeAddress = 0x02 // select mode register
	HMC5883ConfModeValue   = 0x00 // measurement mode
)

type HMC5883Driver struct {
	name       string
	connector  i2c.Connector
	connection i2c.Connection
	Compass    i2c.ThreeDData
	i2c.Config
	gobot.Eventer
}

func NewHMC5883Driver(a i2c.Connector, options ...func(i2c.Config)) *HMC5883Driver {
	m := &HMC5883Driver{
		name:      gobot.DefaultName("HMC5883"),
		connector: a,
		Config:    i2c.NewConfig(),
		Eventer:   gobot.NewEventer(),
	}

	for _, option := range options {
		option(m)
	}

	return m
}

func (d *HMC5883Driver) Name() string {
	return d.name
}

func (d *HMC5883Driver) SetName(s string) {
	d.name = s
}

func (d *HMC5883Driver) Start() error {
	if err := d.initialize(); err != nil {
		return err
	}

	return
}

func (d *HMC5883Driver) Halt() error {
	return nil
}

func (d *HMC5883Driver) Connection() gobot.Connection {
	return d.connection.(gobot.Connection)
}

func (d *HMC5883Driver) initialize() (err error) {
	bus := d.GetBusOrDefault(d.connector.GetDefaultBus())
	address := d.GetAddressOrDefault(HMC5883CompassAddress)

	d.connection, err = d.connector.GetConnection(address, bus)
	if err != nil {
		return err
	}

	// setMode
	if _, err = d.connection.Write([]byte{HMC5883ConfModeAddress, HMC5883ConfModeValue}); err != nil {
		return
	}

	return nil
}

// GetData fetches the latest data from the HMC5883
func (h *HMC5883Driver) GetData() (err error) {
	if _, err = h.connection.Write([]byte{HMC5883DataAddress}); err != nil {
		return
	}

	data := make([]byte, 6)
	_, err = h.connection.Read(data)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	return binary.Read(buf, binary.BigEndian, &h.Compass)
}
