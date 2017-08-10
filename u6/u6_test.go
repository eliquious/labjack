package u6_test

import (
	"bytes"
	"fmt"
	"github.com/eliquious/labjack/u6"
	"github.com/google/gousb"
	"log"
	"testing"
	"time"
)

func Example_Open() {
	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Open U6 connection
	dev, err := u6.OpenUSBConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// log.Println(dev.DeviceDesc())
	fmt.Println(dev.DeviceDesc())
}

func Test_AIN24(t *testing.T) {
	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Open U6 connection
	dev, err := u6.OpenUSBConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// log.Println(dev.DeviceDesc())
	fmt.Println(dev.DeviceDesc())

	ain := &u6.FeedbackAIN24{PositiveChannel: 0, ResolutionIndex: 8, GainIndex: 0, SettlingFactor: 0, Differential: false}
	err = dev.Feedback(ain)
	if err != nil {
		log.Fatal(err)
	}

	voltage, err := ain.GetVoltage()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AIN0: %0.6f\n", voltage)
}

func Test_AIN24Command(t *testing.T) {
	ain := &u6.FeedbackAIN24{PositiveChannel: 0, ResolutionIndex: 0, GainIndex: 0, SettlingFactor: 0, Differential: false}

	var buffer bytes.Buffer
	ain.WriteTo(&buffer)
	t.Log(buffer.Bytes())
}

func Test_BitStateRead(t *testing.T) {
	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Open U6 connection
	dev, err := u6.OpenUSBConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	// log.Println(dev.DeviceDesc())
	fmt.Println(dev.DeviceDesc())

	fio0 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO0}
	fio1 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO1}
	fio2 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO2}
	fio3 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO3}
	fio4 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO4}
	fio6 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO6}
	fio7 := &u6.FeedbackBitStateRead{BitNumber: u6.FIO7}
	err = dev.Feedback(fio0, fio1, fio2, fio3, fio4, fio6, fio7)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("FIO0: %v\n", fio0.GetState())
	fmt.Printf("FIO1: %v\n", fio1.GetState())
	fmt.Printf("FIO2: %v\n", fio2.GetState())
	fmt.Printf("FIO3: %v\n", fio3.GetState())
	fmt.Printf("FIO4: %v\n", fio4.GetState())
	fmt.Printf("FIO6: %v\n", fio6.GetState())
	fmt.Printf("FIO7: %v\n", fio7.GetState())
}

func Test_StreamConfig(t *testing.T) {

	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Open U6 connection
	dev, err := u6.OpenUSBConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	fmt.Println(dev.DeviceDesc())

	_, err = dev.NewStream(&u6.StreamConfig{1, 25, 0, 1000, &u6.ScanConfig{u6.ClockSpeed4Mhz, u6.ClockDivisionOff, 0}, []u6.ChannelConfig{{1, u6.GainIndex1, u6.DifferentialInputDisabled}}})
	if err != nil {
		log.Fatal(err)
	}
}

func Test_StreamData(t *testing.T) {

	// Initialize a new Context.
	ctx := gousb.NewContext()
	defer ctx.Close()
	ctx.Debug(1)

	// Open U6 connection
	dev, err := u6.OpenUSBConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	fmt.Println(dev.DeviceDesc())

	stream, err := dev.NewStream(&u6.StreamConfig{1, 25, 0, 1000, &u6.ScanConfig{u6.ClockSpeed4Mhz, u6.ClockDivisionOff, 0}, []u6.ChannelConfig{{12, u6.GainIndex10, u6.DifferentialInputDisabled}}})
	if err != nil {
		log.Fatal(err)
	}

	ch, err := stream.Start()
	if err != nil {
		log.Fatal(err)
	}

	timeout := time.After(time.Second * 10)
OUTER:
	for {
		select {
		case resp := <-ch:
			fmt.Println("Packet: ", resp.PacketNumber)
			for _, channel := range resp.Data {
				voltage, err := channel.GetCalibratedAIN()
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("\tChannelIndex=%d;\n\tScanNumber=%d;\n\tVoltage=%0.6f\n", channel.ChannelIndex, channel.ScanNumber, voltage)
			}

			// fmt.Println(time.Now(), voltage)
		case <-timeout:
			stream.Stop()
			ch = nil
			break OUTER
		}
	}
}
