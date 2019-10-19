package drivers

import (
	"bytes"
	"encoding/binary"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/i2c"
)

const (
	ITG3200Address         = 0x68
	ITG3200DataAddress     = 0x1B
	ITG3200ConfSMPLAddress = 0x15
	ITG3200ConfSMPLValue   = 0x07 // 每秒采样次数 7+1 次
	ITG3200ConfDLPFAddress = 0x16
	ITG3200ConfDLPFValue   = 0x1E // 000 11 110 FS_SEL=3 DLPF_CFG=6
)

type ITG3200Driver struct {
	name              string
	connector         i2c.Connector
	connection        i2c.Connection
	Gyroscope         ThreeDDataCalibration
	offsetCalibration ThreeDDataCalibration
	Temperature       int16
	i2c.Config
	gobot.Eventer
}

func NewITG3200Driver(a i2c.Connector, options ...func(i2c.Config)) *ITG3200Driver {
	m := &ITG3200Driver{
		name:      gobot.DefaultName("ITG3200"),
		connector: a,
		Config:    i2c.NewConfig(),
		Eventer:   gobot.NewEventer(),
	}

	for _, option := range options {
		option(m)
	}

	return m
}

func (d *ITG3200Driver) Name() string {
	return d.name
}

func (d *ITG3200Driver) SetName(s string) {
	d.name = s
}

func (d *ITG3200Driver) Start() error {
	if err := d.initialize(); err != nil {
		return err
	}

	return nil
}

func (d *ITG3200Driver) Halt() error {
	return nil
}

func (d *ITG3200Driver) Connection() gobot.Connection {
	return d.connection.(gobot.Connection)
}

func (d *ITG3200Driver) initialize() (err error) {
	bus := d.GetBusOrDefault(d.connector.GetDefaultBus())
	address := d.GetAddressOrDefault(ITG3200Address)

	d.connection, err = d.connector.GetConnection(address, bus)
	if err != nil {
		return err
	}

	if _, err = d.connection.Write([]byte{ITG3200ConfSMPLAddress, ITG3200ConfSMPLValue}); err != nil {
		return
	}

	if _, err = d.connection.Write([]byte{ITG3200ConfDLPFAddress, ITG3200ConfDLPFValue}); err != nil {
		return
	}

	return nil
}

func (d *ITG3200Driver) Calibration(times int) {
	var xTotal, yTotal, zTotal int16
	for i := 0; i < times; i++ {
		originGyro, _, err := d.GetRawData()
		if err != nil {
			continue
		}
		xTotal += originGyro.X
		yTotal += originGyro.Y
		zTotal += originGyro.Z
	}

	d.offsetCalibration.X = float64(xTotal) / float64(times)
	d.offsetCalibration.Y = float64(yTotal) / float64(times)
	d.offsetCalibration.Z = float64(zTotal) / float64(times)
}

// GetData fetches the latest data from the ITG3200
func (d *ITG3200Driver) GetRawData() (originGyro i2c.ThreeDData, originTemp int16, err error) {
	if _, err = d.connection.Write([]byte{ITG3200DataAddress}); err != nil {
		return
	}

	data := make([]byte, 8)
	_, err = d.connection.Read(data)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.BigEndian, &originTemp)
	if err != nil {
		return
	}
	err = binary.Read(buf, binary.BigEndian, &originGyro)
	return
}

func (d *ITG3200Driver) GetData() (err error) {
	originGyro, originTemp, err := d.GetRawData()
	if err != nil {
		return
	}

	d.Temperature = originTemp
	d.Gyroscope.X = float64(originGyro.X) - d.offsetCalibration.X
	d.Gyroscope.Y = float64(originGyro.Y) - d.offsetCalibration.Y
	d.Gyroscope.Z = float64(originGyro.Z) - d.offsetCalibration.Z

	return
}
