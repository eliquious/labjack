package u6

import "io"
import "errors"

// DigitalIOBit represents the FIO, EIO and CIO bits
type DigitalIOBit byte

// FIO0 - FIO7, EIO0 - EIO7 and CIO0 - CIO3
const (
	FIO0 DigitalIOBit = iota // 0
	FIO1                     // 1
	FIO2                     // 2
	FIO3                     // 3
	FIO4                     // 4
	FIO5                     // 5
	FIO6                     // 6
	FIO7                     // 7
	EIO0                     // 8
	EIO1                     // 9
	EIO2                     // 10
	EIO3                     // 11
	EIO4                     // 12
	EIO5                     // 13
	EIO6                     // 14
	EIO7                     // 15
	CIO0                     // 16
	CIO1                     // 17
	CIO2                     // 18
	CIO3                     // 19
)

// BitDirection describes the IO direction (read/write)
type BitDirection byte

// BitDirections for BitDirWrite feedback command
const (
	BitDirectionRead  BitDirection = 0   // 0
	BitDirectionWrite BitDirection = 128 // 128
)

// BitState describes the bit state (on/off)
type BitState byte

// BitDirections for BitDirWrite feedback command
const (
	BitStateDisabled BitState = 0   // 0
	BitStateEnabled  BitState = 128 // 128
)

// FeedbackCommand writes to and reads from the USB connection.
type FeedbackCommand interface {
	WriteTo(w io.Writer) (n int, err error)
	ReadFrom(r io.Reader) (n int, err error)
	ResponseSize() int
	SetCalibrationInfo(info CalibrationInfo)
}

// FeedbackPortDirWrite is the Feedback command for PortDirWrite.
type FeedbackPortDirWrite struct {
	FIOWriteMask byte
	EIOWriteMask byte
	CIOWriteMask byte
	FIODirection byte
	EIODirection byte
	CIODirection byte
	calInfo      CalibrationInfo
}

// WriteTo writes the PortDirWrite command.
func (f *FeedbackPortDirWrite) WriteTo(w io.Writer) (int, error) {
	buf := make([]byte, 7)
	buf[0] = 29             // IOType for PortDirWrite
	buf[1] = f.FIOWriteMask //FIO Writemask
	buf[2] = f.EIOWriteMask //EIO Writemask
	buf[3] = f.CIOWriteMask //CIO Writemask
	buf[4] = f.FIODirection //FIO Direction
	buf[5] = f.EIODirection //EIO Direction
	buf[6] = f.CIODirection //CIO Direction
	return w.Write(buf)
}

// SetCalibrationInfo sets the calibration info for calculating the proper values.
func (f *FeedbackPortDirWrite) SetCalibrationInfo(info CalibrationInfo) {
	f.calInfo = info
}

// ReadFrom reads the response.
func (f *FeedbackPortDirWrite) ReadFrom(r io.Reader) (int, error) {
	return 0, nil
}

// ResponseSize returns the response size.
func (f *FeedbackPortDirWrite) ResponseSize() int {
	return 0
}

// FeedbackAIN24 is the Feedback command for AIN24.
type FeedbackAIN24 struct {
	PositiveChannel int
	ResolutionIndex int
	GainIndex       int
	SettlingFactor  int
	Differential    bool
	responseBuffer  []byte
	calInfo         CalibrationInfo
}

// WriteTo writes the FeedbackAIN24 command.
func (f *FeedbackAIN24) WriteTo(w io.Writer) (int, error) {
	// buf[2] = byte((uint(f.ResolutionIndex) & 0x0F) + ((uint(f.GainIndex) & 0x0F) << 4)) // ResolutionIndex + GainInde

	buf := make([]byte, 4)
	buf[0] = 2                                          // IOType for AIN24
	buf[1] = byte(f.PositiveChannel)                    //Positive Channel 0-143s
	buf[2] = byte(uint(f.ResolutionIndex) & 0x0F)       // ResolutionIndex
	buf[2] = byte((uint(f.GainIndex)&0x0F)<<4) + buf[2] // GainIndex
	buf[3] = byte(f.SettlingFactor)                     // SettlingFactor
	if f.Differential {
		buf[3] += 1 << 7
	}
	n, err := w.Write(buf)
	if err != nil {
		return int(n), err
	} else if n != 4 {
		return int(n), errors.New("Feedback AIN24 data was not fully written")
	}
	return int(n), err
}

// ReadFrom reads the response.
func (f *FeedbackAIN24) ReadFrom(r io.Reader) (int, error) {
	f.responseBuffer = make([]byte, 4)
	n, err := r.Read(f.responseBuffer)
	return int(n), err
}

// ResponseSize returns the response size.
func (f *FeedbackAIN24) ResponseSize() int {
	return 3
}

// SetCalibrationInfo sets the calibration info for calculating the proper values.
func (f *FeedbackAIN24) SetCalibrationInfo(info CalibrationInfo) {
	f.calInfo = info
}

// GetVoltage returns the calibrated voltage
func (f *FeedbackAIN24) GetVoltage() (float64, error) {
	return getCalibratedAIN(f.calInfo, f.ResolutionIndex, f.GainIndex, true, uint(f.responseBuffer[0])+uint(f.responseBuffer[1])*256+uint(f.responseBuffer[2])*65536)
}

func getCalibratedAIN(cal CalibrationInfo, ResolutionIndex int, GainIndex int, HiResolution bool, bytesVolt uint) (float64, error) {
	var indexAdjust int
	var analogVolt float64
	// fmt.Printf("Byte volts: %d\n", bytesVolt)
	value := float64(bytesVolt)
	if HiResolution {
		value /= 256.0
	}

	if GainIndex > 4 {
		return 0, errors.New("Invalid gain index")
	}

	if ResolutionIndex > 8 {
		indexAdjust = 24
	}

	if value < cal.CalConstants[indexAdjust+GainIndex*2+9] {
		analogVolt = (cal.CalConstants[indexAdjust+GainIndex*2+9] - value) * cal.CalConstants[indexAdjust+GainIndex*2+8]
	} else {
		analogVolt = (value - cal.CalConstants[indexAdjust+GainIndex*2+9]) * cal.CalConstants[indexAdjust+GainIndex*2]
	}
	return analogVolt, nil
}

// FeedbackBitStateRead is the feedback command for BitStateRead
type FeedbackBitStateRead struct {
	BitNumber DigitalIOBit
	state     byte
}

// WriteTo writes the command
func (f *FeedbackBitStateRead) WriteTo(w io.Writer) (n int, err error) {
	buffer := make([]byte, 2)
	buffer[0] = 10
	buffer[1] = byte(f.BitNumber)
	return w.Write(buffer)
}

// ReadFrom reads the response
func (f *FeedbackBitStateRead) ReadFrom(r io.Reader) (n int, err error) {
	responseBuffer := make([]byte, 1)
	n, err = r.Read(responseBuffer)
	f.state = responseBuffer[0]
	return n, err
}

// ResponseSize is the size of the response
func (f *FeedbackBitStateRead) ResponseSize() int {
	return 1
}

// SetCalibrationInfo sets the calibration info
func (f *FeedbackBitStateRead) SetCalibrationInfo(info CalibrationInfo) {
}

// GetState gets the response state of the bit.
func (f *FeedbackBitStateRead) GetState() bool {
	return f.state == byte(1)
}

// FeedbackBitDirWrite is the BitDirWrite feedback command
type FeedbackBitDirWrite struct {
	BitNumber DigitalIOBit
	Direction BitDirection
	calInfo   CalibrationInfo
	state     byte
}

// WriteTo writes the command
func (f *FeedbackBitDirWrite) WriteTo(w io.Writer) (n int, err error) {
	buffer := make([]byte, 2)
	buffer[0] = 13
	buffer[1] = byte(f.BitNumber) + byte(f.Direction)
	return w.Write(buffer)
}

// ReadFrom reads the response
func (f *FeedbackBitDirWrite) ReadFrom(r io.Reader) (n int, err error) {
	responseBuffer := make([]byte, 1)
	n, err = r.Read(responseBuffer)
	f.state = responseBuffer[0]
	return n, err
}

// ResponseSize returns the size of the response
func (f *FeedbackBitDirWrite) ResponseSize() int {
	return 1
}

// SetCalibrationInfo sets the CalibrationInfo
func (f *FeedbackBitDirWrite) SetCalibrationInfo(info CalibrationInfo) {
}

// FeedbackBitStateWrite is the BitStateWrite feedback command
type FeedbackBitStateWrite struct {
	BitNumber DigitalIOBit
	State     BitState
}

// WriteTo writes the command
func (f *FeedbackBitStateWrite) WriteTo(w io.Writer) (n int, err error) {
	buffer := make([]byte, 2)
	buffer[0] = 11
	buffer[1] = byte(f.BitNumber) + byte(f.State)
	return w.Write(buffer)
}

// ReadFrom reads the response
func (f *FeedbackBitStateWrite) ReadFrom(r io.Reader) (n int, err error) {
	return 0, nil
}

// ResponseSize returns the size of the response
func (f *FeedbackBitStateWrite) ResponseSize() int {
	return 0
}

// SetCalibrationInfo sets the CalibrationInfo
func (f *FeedbackBitStateWrite) SetCalibrationInfo(info CalibrationInfo) {
}
