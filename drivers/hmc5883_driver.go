package drivers

import (
	"bytes"
	"encoding/binary"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
	"math"
)

const (
	HMC5883CompassAddress  = 0x1E
	HMC5883DataAddress     = 0x03
	HMC5883ConfAAddress    = 0x00 // register A
	HMC5883ConfAValue      = 0x70
	HMC5883ConfModeAddress = 0x02 // select mode register
	HMC5883ConfModeValue   = 0x00 // continues measurement mode
)

type HMC5883Driver struct {
	name                string
	connector           i2c.Connector
	connection          i2c.Connection
	Compass             i2c.ThreeDData
	magneticDeclination float64 // 地磁偏角
	i2c.Config
	gobot.Eventer
}

func NewHMC5883Driver(a i2c.Connector, magneticDeclination float64, options ...func(i2c.Config)) *HMC5883Driver {
	m := &HMC5883Driver{
		name:                gobot.DefaultName("HMC5883"),
		connector:           a,
		magneticDeclination: magneticDeclination,
		Config:              i2c.NewConfig(),
		Eventer:             gobot.NewEventer(),
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

	return nil
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

	// setA
	if _, err = d.connection.Write([]byte{HMC5883ConfAAddress, HMC5883ConfAValue}); err != nil {
		return
	}

	// setMode
	if _, err = d.connection.Write([]byte{HMC5883ConfModeAddress, HMC5883ConfModeValue}); err != nil {
		return
	}

	return nil
}

// GetData fetches the latest data from the HMC5883
func (d *HMC5883Driver) GetData() (err error) {
	if _, err = d.connection.Write([]byte{HMC5883DataAddress}); err != nil {
		return
	}

	data := make([]byte, 6)
	_, err = d.connection.Read(data)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	return binary.Read(buf, binary.BigEndian, &d.Compass)
}

func (d *HMC5883Driver) Heading() float64 {
	radians := math.Atan2(float64(d.Compass.Y), float64(d.Compass.X))
	if radians < 0 {
		radians += 2 * math.Pi
	}

	degrees := (radians * 180 / math.Pi) + d.magneticDeclination
	if degrees > 360 {
		degrees -= 360
	}

	return degrees
}
