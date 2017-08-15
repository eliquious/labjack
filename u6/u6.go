package u6

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/eliquious/labjack"
	"github.com/google/gousb"
	// "io"
)

// OpenUSBConnection opens the USB connection a LabJack U6.
func OpenUSBConnection(usbctx *gousb.Context) (*U6, error) {
	if usbctx == nil {
		return &emptyU6, ErrInvalidContext
	}

	// Open any device with a given VID/PID using a convenience function.
	dev, err := usbctx.OpenDeviceWithVIDPID(labjack.LabJackVendorID, labjack.U6ProductID)
	if err != nil {
		return &emptyU6, ErrLibUSB{"Could not open a device", err}
	}
	if err := dev.Reset(); err != nil {
		return &emptyU6, err
	}
	if err := dev.SetAutoDetach(true); err != nil {
		return &emptyU6, err
	}

	ljdev := &U6{dev, DeviceDesc{}, DefaultCalibrationInfo}
	if err = ljdev.initConnection(); err != nil {
		return &emptyU6, err
	}

	if err = ljdev.getCalibrationInfo(); err != nil {
		return &emptyU6, err
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

var feedbackHeader = []byte{0, 0xF8, 0, 0, 0, 0, 0}

// Feedback executes all of the Feedback commands given.
func (u *U6) Feedback(cmds ...FeedbackCommand) error {
	var sendBuffer bytes.Buffer

	// Write header
	n, err := sendBuffer.Write(feedbackHeader)
	if err != nil {
		return err
	} else if n != len(feedbackHeader) {
		return errors.New("Feedback header could not be written")
	}
	// fmt.Println("After header: ", sendBuffer.Bytes())

	// Write each feeback command
	var length int
	var responseSize int
	for _, cmd := range cmds {
		cmd.SetCalibrationInfo(u.calibration)
		n, err := cmd.WriteTo(&sendBuffer)
		if err != nil {
			return err
		} else if n == 0 {
			return errors.New("Command data was not written")
		}
		length += n
		responseSize += cmd.ResponseSize()
	}
	// fmt.Println("After commands: ", sendBuffer.Bytes())

	// Pad message if needed
	if (length+1)%2 == 1 {
		if err = sendBuffer.WriteByte(0x00); err != nil {
			return err
		}
		length++
	}
	// fmt.Println("After padding: ", sendBuffer.Bytes())

	// Get bytes and set word count
	buf := sendBuffer.Bytes()
	buf[2] = byte(length+1) / 2
	// fmt.Println("After word count: ", sendBuffer.Bytes())

	// Calculate checksum
	if err = setChecksum(buf); err != nil {
		return err
	}
	fmt.Printf("After checksum: %v\n", buf)

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
	n, err = out.Write(buf)
	if err != nil {
		return err
	} else if n != len(buf) {
		return errors.New("Send buffer was not written completely")
	}

	// Open endpoint
	in, err := inf.InEndpoint(labjack.U6PipeInEP2)
	if err != nil {
		return err
	}

	// Read response
	recvBuffer := make([]byte, 9+responseSize)
	n, err = in.Read(recvBuffer)
	if err != nil {
		return err
	} else if n != len(recvBuffer) {
		// fmt.Printf("Recv buffer: %v\n", recvBuffer)
		return fmt.Errorf("Full response was not recieved from device: %d != %d", n, len(recvBuffer))
	}

	checksumTotal, err := extendedChecksum16(recvBuffer)
	if err != nil {
		return err
	} else if byte((checksumTotal/256)&0xFF) != recvBuffer[5] {
		return ErrInvalidChecksumResponse
	} else if byte(checksumTotal&0xFF) != recvBuffer[4] {
		return ErrInvalidChecksumResponse
	} else if recvBuffer[1] != 0xF8 || recvBuffer[3] != 0x00 {
		return ErrInvalidResponseHeader
	} else if recvBuffer[6] != 0x00 {
		return ErrInvalidResponseHeader
	}

	c8, err := extendedChecksum8(recvBuffer)
	if err != nil {
		return err
	} else if c8 != recvBuffer[0] {
		return ErrInvalidResponseHeader
	}
	// fmt.Println("Recv Buffer: ", recvBuffer)

	errCode := recvBuffer[6]
	errFrame := recvBuffer[7]
	if errCode != 0 {
		return fmt.Errorf("Feedback response error code (%d): command=%d", errCode, errFrame)
	}

	// Populate the commands' response
	remaining := int64(len(recvBuffer) - 9)
	buffer := bytes.NewBuffer(recvBuffer[9:])
	for _, cmd := range cmds {

		num, err := cmd.ReadFrom(buffer)
		if err != nil {
			return err
		}
		remaining -= int64(num)
	}

	if remaining != 0 {
		return fmt.Errorf("Feedback response was not decoded completely: remaining=%d", remaining)
	}
	return nil
}

// NewStream creates a new data stream
func (u *U6) NewStream(config *StreamConfig) (*Stream, error) {
	stream := &Stream{u, config, make(chan struct{}, 1), func() {}}
	if config.SamplesPerPacket < 1 || config.SamplesPerPacket > 25 {
		return stream, errors.New("Invalid samples per packet")
	} else if config.ResolutionIndex < 1 || config.ResolutionIndex > 8 {
		return stream, errors.New("Invalid resolution index")
	}

	// config.ScanFrequency *= len(config.Channels)
	if config.ScanFrequency != 0 {
		if config.ScanFrequency < 1000 {
			if config.ScanFrequency < 25 {
				config.SamplesPerPacket = byte(config.ScanFrequency)
			}
			config.ScanConfig.DivideBy256 = ClockDivisionOn
			config.ScanConfig.ScanInterval = uint16(15625 / config.ScanFrequency)
		} else {
			config.ScanConfig.DivideBy256 = ClockDivisionOff
			config.ScanConfig.ScanInterval = uint16(4000000 / config.ScanFrequency)
		}
	}

	if config.SamplesPerPacket > 25 {
		config.SamplesPerPacket = 25
	} else if config.SamplesPerPacket < 1 {
		config.SamplesPerPacket = 1
	}

	header := make([]byte, 14+2*len(config.Channels))
	header[1] = 0xF8
	header[2] = byte(len(config.Channels) + 4)
	header[3] = 0x11
	header[6] = byte(len(config.Channels))
	header[7] = byte(config.ResolutionIndex)
	header[8] = byte(config.SamplesPerPacket)
	header[10] = byte(config.SettlingFactor)
	header[11] = config.ScanConfig.GetByte()

	// // scanInterval := 4000
	// if config.ScanConfig.ClockSpeed == ClockSpeed48Mhz {
	// 	// scanInterval = 48000
	// }
	header[12] = byte(config.ScanConfig.ScanInterval & 0x00FF)
	header[13] = byte(config.ScanConfig.ScanInterval / 256)

	for i, ch := range config.Channels {
		header[14+i*2] = ch.PositiveChannel
		header[15+i*2] = byte(ch.Differential) + byte(ch.GainIndex)<<4
	}
	// fmt.Println("After header: ", sendBuffer.Bytes())

	// Calculate checksum
	if err := setChecksum(header); err != nil {
		return stream, err
	}
	// fmt.Printf("After checksum: %v\n", header)

	// Open USB interface
	inf, done, err := u.device.DefaultInterface()
	if err != nil {
		return stream, err
	}
	defer done()

	// Open endpoint
	out, err := inf.OutEndpoint(labjack.U6PipeOutEP1)
	if err != nil {
		return stream, err
	}

	// Transmit send buffer
	n, err := out.Write(header)
	if err != nil {
		return stream, err
	} else if n != len(header) {
		return stream, errors.New("Send buffer was not written completely")
	}

	// Open endpoint
	in, err := inf.InEndpoint(labjack.U6PipeInEP2)
	if err != nil {
		return stream, err
	}

	// Read response
	recvBuffer := make([]byte, 8)
	n, err = in.Read(recvBuffer)
	if err != nil {
		return stream, err
	} else if n != len(recvBuffer) {
		// fmt.Printf("Recv buffer: %v\n", recvBuffer)
		return stream, fmt.Errorf("Full response was not recieved from device: %d != %d", n, len(recvBuffer))
	}

	checksumTotal, err := extendedChecksum16(recvBuffer)
	if err != nil {
		return stream, err
	} else if byte((checksumTotal/256)&0xFF) != recvBuffer[5] {
		return stream, ErrInvalidChecksumResponse
	} else if byte(checksumTotal&0xFF) != recvBuffer[4] {
		return stream, ErrInvalidChecksumResponse
	} else if recvBuffer[1] != 0xF8 || recvBuffer[3] != 0x11 {
		return stream, ErrInvalidResponseHeader
	} else if recvBuffer[7] != 0x00 {
		return stream, ErrInvalidResponseHeader
	}

	c8, err := extendedChecksum8(recvBuffer)
	if err != nil {
		return stream, err
	} else if c8 != recvBuffer[0] {
		return stream, ErrInvalidResponseHeader
	}
	// fmt.Println("Recv Buffer: ", recvBuffer)

	errCode := recvBuffer[6]
	if errCode != 0 {
		return stream, fmt.Errorf("Feedback response error code (%d)", errCode)
	}
	return stream, nil
}
