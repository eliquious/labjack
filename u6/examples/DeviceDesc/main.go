package main

import (
	"fmt"
	"github.com/eliquious/labjack/u6"
	"github.com/google/gousb"
	"log"
	"time"
)

func main() {
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
	fmt.Println(dev.GetCalibrationInfo())

	start := time.Now()
	// ticker := time.Tick(time.Millisecond * 5)
	ain := &u6.FeedbackAIN24{PositiveChannel: 0, ResolutionIndex: 8, GainIndex: 0, SettlingFactor: 0, Differential: false}
	for i := 0; i < 1000; i++ {

		err = dev.Feedback(ain)
		if err != nil {
			log.Fatal(err)
		}

		_, err := ain.GetVoltage()
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(time.Since(start))
	fmt.Println(time.Since(start) / 1000)
}
