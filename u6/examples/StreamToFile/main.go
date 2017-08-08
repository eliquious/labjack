package main

import (
	"fmt"
	"github.com/eliquious/labjack/u6"
	"github.com/google/gousb"
	"log"
	"os"
	"time"
)

func main() {

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

	stream, err := dev.NewStream(u6.StreamConfig{1, 25, 0, u6.ScanConfig{u6.ClockSpeed4Mhz, u6.ClockDivisionOff}, []u6.ChannelConfig{{12, u6.GainIndex10, u6.DifferentialInputDisabled}}})
	if err != nil {
		log.Fatal(err)
	}

	ch, err := stream.Start()
	if err != nil {
		log.Fatal(err)
	}

	fh, _ := os.Create("out.csv")
	defer fh.Close()

	timeout := time.After(time.Second * 30)
OUTER:
	for {
		select {
		case resp := <-ch:
			// fmt.Println("Packet: ", resp.PacketNumber)
			for _, channel := range resp.Data {
				voltage, err := channel.GetCalibratedAIN()
				if err != nil {
					fmt.Println(err)
					continue
				}
				fh.WriteString(fmt.Sprintf("%s,%d,%d,%d,%0.8f\n", time.Now().Format(time.RFC3339Nano), channel.ChannelIndex, channel.ScanNumber, resp.PacketNumber, voltage))
				fmt.Printf("Packet=%d; ChannelIndex=%d; ScanNumber=%d; Voltage=%0.6f\n", resp.PacketNumber, channel.ChannelIndex, channel.ScanNumber, voltage)
			}

			// fmt.Println(time.Now(), voltage)
		case <-timeout:
			stream.Stop()
			ch = nil
			break OUTER
		}
	}
}
