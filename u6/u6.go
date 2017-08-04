package u6

import (
	"fmt"
	"github.com/eliquious/labjack"
	"github.com/google/gousb"
)

// OpenUSBConnection opens the USB connection a LabJack U6.
func OpenUSBConnection(usbctx *gousb.Context) (U6, error) {
	if usbctx == nil {
		return emptyU6, ErrInvalidContext
	}

	// Open any device with a given VID/PID using a convenience function.
	dev, err := usbctx.OpenDeviceWithVIDPID(labjack.LabJackVendorID, labjack.U6ProductID)
	if err != nil {
		return emptyU6, ErrLibUSB{"Could not open a device", err}
	}

	ljdev := U6{dev, DeviceDesc{}, DefaultCalibrationInfo}
	if err = ljdev.initConnection(); err != nil {
		return emptyU6, err
	}

	if err = ljdev.getCalibrationInfo(); err != nil {
		return emptyU6, err
	}
	return ljdev, nil
}

var emptyU6 U6

// U6 represents the LabJack U6 / U6 Pro devices
type U6 struct {
	device      *gousb.Device
	config      DeviceDesc
	calibration CalibrationInfo
}

// DeviceDesc returns the device details.
func (u *U6) DeviceDesc() DeviceDesc {
	return u.config
}

func (u *U6) initConnection() error {
	sendBuffer := make([]byte, 26)
	recBuffer := make([]byte, 38)

	// setting up U6Config
	sendBuffer[1] = uint8(0xF8)
	sendBuffer[2] = uint8(0x0A)
	sendBuffer[3] = uint8(0x08)

	for i := 6; i < 26; i++ {
		sendBuffer[i] = uint8(0x00)
	}
	extendedChecksum(sendBuffer)

	// Open USB interface
	inf, done, err := u.device.DefaultInterface()
	if err != nil {
		return err
	}
	defer done()

	// Open endpoint
	out, err := inf.OutEndpoint(labjack.U6PipeOutEP1)
	if err != nil {
		return err
	}

	// Transmit send buffer
	n, err := out.Write(sendBuffer)
	if err != nil {
		return err
	} else if n != len(sendBuffer) {
		return ErrEndpointSendError
	}

	// Open endpoint
	in, err := inf.InEndpoint(labjack.U6PipeInEP2)
	if err != nil {
		return err
	}

	// Read response
	n, err = in.Read(recBuffer)
	if err != nil {
		return err
	} else if n != len(recBuffer) {
		return ErrEndpointRecvError
	}

	// Validate response
	if err = validateCommandResponse(recBuffer); err != nil {
		return err
	}

	// Parse device info
	config, err := parseConfigBytes(recBuffer)
	if err != nil {
		return err
	}
	u.config = config

	return nil
}

func validateCommandResponse(recBuffer []uint8) error {
	if len(recBuffer) < 7 {
		return ErrResponseTooShort
	}

	// Bad checksum response
	if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
		return ErrInvalidChecksumResponse
	}

	// Validate response header
	if recBuffer[1] != uint8(0xF8) || recBuffer[2] != uint8(0x10) ||
		recBuffer[3] != uint8(0x08) {
		return ErrInvalidResponseHeader
	} else if recBuffer[6] != 0 {
		return ErrLabJackErrorCode{int(recBuffer[6])}
	}

	// Validate response
	checksumTotal, err := extendedChecksum16(recBuffer)
	if err != nil {
		return err
	} else if uint8((checksumTotal/256)&0xFF) != recBuffer[5] {
		fmt.Printf("ErrInvalidChecksum MSB: %s != %s\n", uint8((checksumTotal/256)&0xFF), recBuffer[5])
		return ErrInvalidChecksum
	} else if uint8(checksumTotal&0xFF) != recBuffer[4] {
		fmt.Printf("ErrInvalidChecksum LSB: %s != %s\n", uint8((checksumTotal)&0xFF), recBuffer[4])
		return ErrInvalidChecksum
	}

	c, err := extendedChecksum8(recBuffer)
	if err != nil {
		return err
	} else if c != recBuffer[0] {
		fmt.Printf("ErrInvalidChecksum 8-bit: %d != %d\n", c, recBuffer[0])
		return ErrInvalidChecksum
	}

	return nil
}

// GetCalibrationInfo gets the calibration information for the device
func (u *U6) getCalibrationInfo() error {
	sendBuffer := make([]byte, 64)
	recBuffer := make([]byte, 64)

	// setting up U6Config
	sendBuffer[1] = uint8(0xF8)
	sendBuffer[2] = uint8(0x0A)
	sendBuffer[3] = uint8(0x08)

	for i := 6; i < 26; i++ {
		sendBuffer[i] = uint8(0x00)
	}
	extendedChecksum(sendBuffer[:26])

	// Open USB interface
	inf, done, err := u.device.DefaultInterface()
	if err != nil {
		return err
	}
	defer done()

	// Open endpoint
	out, err := inf.OutEndpoint(labjack.U6PipeOutEP1)
	if err != nil {
		return err
	}

	// Transmit send buffer
	n, err := out.Write(sendBuffer[:26])
	if err != nil {
		return err
	} else if n != 26 {
		return ErrEndpointSendError
	}

	// Open endpoint
	in, err := inf.InEndpoint(labjack.U6PipeInEP2)
	if err != nil {
		return err
	}

	// Read response
	n, err = in.Read(recBuffer[:38])
	if err != nil {
		return err
	} else if n != 38 {
		return ErrEndpointRecvError
	}

	// Validate response
	if err = validateCommandResponse(recBuffer); err != nil {
		return err
	}

	cal := CalibrationInfo{
		ProductID:    6,
		HiResolution: (recBuffer[37] & 8) == 8,
	}

	var offset int
	for i := 0; i < 10; i++ {

		/* reading block i from memory */
		sendBuffer[1] = uint8(0xF8) //command byte
		sendBuffer[2] = uint8(0x01) //number of data words
		sendBuffer[3] = uint8(0x2D) //extended command number
		sendBuffer[6] = 0
		sendBuffer[7] = uint8(i) //Blocknum = i
		// extendedChecksum(sendBuffer[:8])
		setChecksum(sendBuffer[:8])
		// fmt.Println("Sent: ", sendBuffer[:8])

		// Transmit send buffer
		n, err := out.Write(sendBuffer[:8])
		if err != nil {
			return err
		} else if n != 8 {
			return ErrEndpointSendError
		}

		// Read response
		n, err = in.Read(recBuffer[:40])
		if err != nil {
			return err
		} else if recBuffer[0] == 0xB8 && recBuffer[1] == 0xB8 {
			// Bad checksum response
			return ErrInvalidChecksumResponse
		} else if n != 40 {
			// fmt.Printf("Error reading response: n=%d; i=%d; offset=%d\n", n, i, offset)
			return ErrEndpointRecvError
		}
		// fmt.Println("Recv'd: ", recBuffer[:40])

		if recBuffer[1] != uint8(0xF8) || recBuffer[2] != uint8(0x11) || recBuffer[3] != uint8(0x2D) {
			return ErrInvalidResponseHeader
		}
		offset = i * 4

		//block data starts on byte 8 of the buffer
		cal.CalConstants[offset] = uint8ArrayToFloat64(recBuffer[8:], 0)
		// fmt.Println(recBuffer[8:16])
		cal.CalConstants[offset+1] = uint8ArrayToFloat64(recBuffer[8:], 8)
		// fmt.Println(recBuffer[16:24])
		cal.CalConstants[offset+2] = uint8ArrayToFloat64(recBuffer[8:], 16)
		// fmt.Println(recBuffer[24:32])
		cal.CalConstants[offset+3] = uint8ArrayToFloat64(recBuffer[8:], 24)
		// fmt.Println(recBuffer[32:40])
	}
	u.calibration = cal

	return nil
}

// GetCalibrationInfo gets the calibration information for the device
func (u *U6) GetCalibrationInfo() CalibrationInfo {
	return u.calibration
}

// Close closes the device connection.
func (u *U6) Close() error {
	return u.device.Close()
}
