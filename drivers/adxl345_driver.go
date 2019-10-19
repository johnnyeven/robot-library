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
	Adxl345GravityLSB       = 256
)

type ADXL345Driver struct {
	name              string
	connector         i2c.Connector
	connection        i2c.Connection
	Accelerometer     ThreeDDataCalibration
	offsetCalibration ThreeDDataCalibration
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

func (d *ADXL345Driver) Calibration(times int) {
	var xTotal, yTotal, zTotal int16
	for i := 0; i < times; i++ {
		origin, err := d.GetRawData()
		if err != nil {
			continue
		}
		xTotal += origin.X
		yTotal += origin.Y
		zTotal += origin.Z
	}

	d.offsetCalibration.X = float64(xTotal) / float64(times)
	d.offsetCalibration.Y = float64(yTotal) / float64(times)
	d.offsetCalibration.Z = float64(zTotal)/float64(times) + Adxl345GravityLSB // 抵消重力加速度1g
}

// GetData fetches the latest data from the ADXL345
func (d *ADXL345Driver) GetRawData() (origin i2c.ThreeDData, err error) {
	if _, err = d.connection.Write([]byte{Adxl345DataAddress}); err != nil {
		return
	}

	data := make([]byte, 6)
	_, err = d.connection.Read(data)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.BigEndian, &origin)

	return
}

// GetData fetches the latest data from the ADXL345
func (d *ADXL345Driver) GetData() (err error) {
	origin, err := d.GetRawData()
	if err != nil {
		return
	}

	d.Accelerometer.X = float64(origin.X) - d.offsetCalibration.X
	d.Accelerometer.Y = float64(origin.Y) - d.offsetCalibration.Y
	d.Accelerometer.Z = float64(origin.Z) - d.offsetCalibration.Z

	return
}
