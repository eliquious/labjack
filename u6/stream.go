package u6

type StreamConfig struct {
	ResolutionIndex  byte
	SamplesPerPacket byte
	SettlingFactor   byte
	ScanConfig       ScanConfig
	Channels         []ChannelConfig
}

type ChannelConfig struct {
	PositiveChannel byte
	GainIndex       GainIndex
	Differential    DifferentialInput
}

type ScanConfig struct {
	ClockSpeed  ClockSpeed
	DivideBy256 ClockDivision
}

func (s ScanConfig) GetByte() byte {
	return byte(s.ClockSpeed) + byte(s.DivideBy256)
}

type ClockSpeed byte

const (
	ClockSpeed4Mhz  ClockSpeed = 0
	ClockSpeed48Mhz ClockSpeed = 8
)

type ClockDivision byte

const (
	ClockDivisionOff ClockDivision = 0
	ClockDivisionOn  ClockDivision = 2
)

type DifferentialInput byte

const (
	DifferentialInputDisabled DifferentialInput = 0   // 0
	DifferentialInputEnabled  DifferentialInput = 128 // 128
)

type GainIndex byte

const (
	GainIndex1 GainIndex = iota
	GainIndex10
	GainIndex100
	GainIndex1000
)

type Stream struct {
	device *U6
	config StreamConfig
}

func (s *Stream) Start() error {

	header := make([]byte, 2)
	header[0] = 0xA8
	header[1] = 0xA8
	// fmt.Println("After header: ", sendBuffer.Bytes())

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
	n, err := out.Write(header)
	if err != nil {
		return err
	} else if n != len(header) {
		return errors.New("Send buffer was not written completely")
	}

	// Open endpoint
	in, err := inf.InEndpoint(labjack.U6PipeInEP2)
	if err != nil {
		return err
	}

	// Read response
	recvBuffer := make([]byte, 4)
	n, err = in.Read(recvBuffer)
	if err != nil {
		return err
	} else if n != len(recvBuffer) {
		// fmt.Printf("Recv buffer: %v\n", recvBuffer)
		return fmt.Errorf("Full response was not recieved from device: %d != %d", n, len(recvBuffer))
	}

	if normalChecksum8(recvBuffer) != recvBuffer[0] {
		return ErrInvalidChecksumResponse
	} else if recvBuffer[1] != 0xA8 || recvBuffer[3] != 0x00 {
		return ErrInvalidResponseHeader
	}

	errCode := recvBuffer[2]
	if errCode != 0 {
		return fmt.Errorf("Feedback response error code (%d)", errCode)
	}
	return nil
}

func (s *Stream) Stop() {
}
