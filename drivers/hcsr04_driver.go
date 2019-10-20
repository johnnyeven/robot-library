package drivers

import (
	"fmt"
	"gobot.io/x/gobot"
	"time"
)

type HCSR04Driver struct {
	name    string
	trigPin string
	echoPin string

	connection Connection
}

func NewHCSR04Driver(a Connection, trigPin string, echoPin string) *HCSR04Driver {
	return &HCSR04Driver{
		name:       gobot.DefaultName("HCSR04"),
		trigPin:    trigPin,
		echoPin:    echoPin,
		connection: a,
	}
}

func (d *HCSR04Driver) Name() string {
	return d.name
}

func (d *HCSR04Driver) SetName(s string) {
	d.name = s
}

func (d *HCSR04Driver) Start() error {
	return nil
}

func (d *HCSR04Driver) Halt() error {
	return nil
}

func (d *HCSR04Driver) Connection() gobot.Connection {
	return d.connection.(gobot.Connection)
}

func (d *HCSR04Driver) Measure() (float64, error) {
	timer := time.NewTimer(5 * time.Second)
	// 持续输出一个 10微秒的高电平启动超声波发射
	err := d.connection.DigitalWrite(d.trigPin, 1)
	if err != nil {
		return 0, err
	}
	time.Sleep(10 * time.Microsecond)
	err = d.connection.DigitalWrite(d.trigPin, 0)
	if err != nil {
		return 0, err
	}

	// 等待输出端变为高电平，之后记录起始时间
Start:
	for {
		select {
		case <-timer.C:
			return 0, fmt.Errorf("[HCSR04Driver] Measure start err: timeout")
		default:
			echo, err := d.connection.DigitalRead(d.echoPin)
			if err != nil {
				return 0, err
			}
			if echo > 0 {
				timer.Reset(5 * time.Second)
				break Start
			}
		}
	}
	start := time.Now()

	// 等待输出端变为低电平，之后记录结束时间
Wait:
	for {
		select {
		case <-timer.C:
			return 0, fmt.Errorf("[HCSR04Driver] Measure wait err: timeout")
		default:
			echo, err := d.connection.DigitalRead(d.echoPin)
			if err != nil {
				return 0, err
			}
			if echo == 0 {
				break Wait
			}
		}
	}
	end := time.Now()

	/*
		计算距离：
		    距离(单位:m)
		                = (start - end) * 声波速度 / 2
		    声波速度取 343m/s 。
		    然后再把测得的距离转换为 cm。
		    距离(单位:cm)
		                = (start - end) * 声波速度 / 2 * 100
		                = (start - end) * 17150
	*/
	offset := end.Sub(start)
	distance := offset.Seconds() * 17150
	return distance, nil
}
