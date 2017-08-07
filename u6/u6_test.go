package u6_test

import (
	"bytes"
	"fmt"
	"github.com/eliquious/labjack/u6"
	"github.com/google/gousb"
	"log"
	"testing"
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

func Test_AIN24ARCommand(t *testing.T) {
	ain := &u6.FeedbackAIN24{PositiveChannel: 0, ResolutionIndex: 0, GainIndex: 0, SettlingFactor: 0, Differential: false}

	var buffer bytes.Buffer
	ain.WriteTo(&buffer)
	t.Log(buffer.Bytes())
}

// func Test_getCalibratedAIN(t *testing.T) {
// 	value, err := getCalibratedAIN(DefaultCalibrationInfo, 8, 0, false)
// }
