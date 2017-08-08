package u6

import (
	// "bufio"
	"errors"
	"fmt"
	"github.com/eliquious/labjack"
	"github.com/google/gousb"
	"io"
	"time"
)

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

type StreamResponse struct {
	Timestamp    time.Time
	Data         []ChannelData
	PacketNumber int
	ErrorCode    byte
	Error        error
}

// func (s StreamResponse) String() string {
// 	return fmt.Sprintf("<u6.>", s.PacketNumber, s.Data)
// }

type ChannelData struct {
	Raw           uint16
	ChannelIndex  int
	ScanNumber    int
	PacketNumber  int
	config        StreamConfig
	calInfo       CalibrationInfo
	channelConfig ChannelConfig
}

func (c *ChannelData) GetCalibratedAIN() (float64, error) {
	return getCalibratedAIN(c.calInfo, int(c.config.ResolutionIndex), int(c.channelConfig.GainIndex), false, uint(c.Raw))
}

type Stream struct {
	device   *U6
	config   StreamConfig
	stopCh   chan struct{}
	closeInf func()
}

func (s *Stream) Start() (chan StreamResponse, error) {
	// var dataCh chan StreamResponse
	dataCh := make(chan StreamResponse, 100)

	// Stop existing streams
	err := s.stop()
	if err != nil {
		return dataCh, err
	}
	// fmt.Println("Stopped any existing streams")

	// Start new stream
	err = s.start()
	if err != nil {
		return dataCh, err
	}
	// fmt.Println("Started new stream")

	// Open USB interface
	inf, done, err := s.device.device.DefaultInterface()
	if err != nil {
		return dataCh, err
	}
	// defer done()

	// Open endpoint
	in, err := inf.InEndpoint(labjack.U6PipeInEP3)
	if err != nil {
		done()
		return dataCh, err
	}
	// in.Timeout = time.Second

	stream, err := in.NewStream(int(14*s.config.SamplesPerPacket*2), 20)
	if err != nil {
		done()
		return dataCh, err
	}
	s.closeInf = done

	fmt.Println("Reading stream")
	go s.readStream(dataCh, stream)
	return dataCh, nil
}

func (s *Stream) readStream(dataCh chan StreamResponse, stream *gousb.ReadStream) {
	defer stream.Close()

	var n int
	var err error
	// var packetID byte
	var scanNumber int
	var channelIndex int
	var packetNumber int
	var checksumTotal8 uint8
	var checksumTotal16 uint16
	samplesPerPacket := s.config.SamplesPerPacket
	bytelimit := int(12 + s.config.SamplesPerPacket*2)
	numChannels := len(s.config.Channels)
	recvBuffer := make([]byte, 14+s.config.SamplesPerPacket*2)
	// bufferedReader := bufio.NewReader(stream)
	for {
		select {
		case <-s.stopCh:
			s.closeInf()
			s.stop()
			return
		default:
			// n, err = bufferedReader.Read(recvBuffer)
			n, err = io.ReadFull(stream, recvBuffer)
			// fmt.Println(err)
			if err != nil {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: err}
				continue
			} else if n != len(recvBuffer) {
				fmt.Printf("Failed to read complete response: %d != %d\n", n, len(recvBuffer))
				// fmt.Println(recvBuffer)
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrResponseTooShort}
				continue
			}
			// fmt.Println(recvBuffer)

			checksumTotal16, err = extendedChecksum16(recvBuffer)
			if err != nil {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: err}
				continue
			} else if byte((checksumTotal16>>8)&0xff) != recvBuffer[5] {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrInvalidChecksumResponse}
				continue
			} else if byte(checksumTotal16&0xff) != recvBuffer[4] {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrInvalidChecksumResponse}
				continue
			}

			checksumTotal8, err = extendedChecksum8(recvBuffer)
			if err != nil {
				dataCh <- StreamResponse{Error: err}
				continue
			} else if checksumTotal8 != recvBuffer[0] {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrInvalidChecksumResponse}
				continue
			}

			if recvBuffer[1] != byte(0xF9) {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrInvalidResponseHeader}
				continue
			} else if recvBuffer[2] != byte(4+s.config.SamplesPerPacket) {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrInvalidResponseHeader}
				continue
			} else if recvBuffer[3] != byte(0xC0) {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrInvalidResponseHeader}
				continue
			}

			if recvBuffer[11] == byte(59) {
				// Data overflow
			} else if recvBuffer[11] == byte(60) {
				// Auto-recovery packet
				// recvBuffer[6] + recvBuffer[7]*256 scans dropped
			} else if recvBuffer[11] != 0 {
				dataCh <- StreamResponse{Timestamp: time.Now(), Error: ErrLabJackErrorCode{int(recvBuffer[11])}}
				continue
			}

			// Packet ID
			// packetID = recvBuffer[10]

			// Backlog
			// Channel data
			data := make([]ChannelData, samplesPerPacket)
			for i := 12; i < bytelimit; i += 2 {
				data[(i-12)/2] = ChannelData{
					ChannelIndex:  channelIndex,
					ScanNumber:    scanNumber,
					PacketNumber:  packetNumber,
					Raw:           uint16(recvBuffer[i]) + uint16(recvBuffer[i+1])*256,
					calInfo:       s.device.calibration,
					config:        s.config,
					channelConfig: s.config.Channels[channelIndex],
				}
				channelIndex++
				if channelIndex >= numChannels {
					channelIndex = 0
					scanNumber++
				}
			}

			if packetNumber >= 255 {
				packetNumber = 0
			}
			packetNumber++
			dataCh <- StreamResponse{Timestamp: time.Now(), Data: data, PacketNumber: packetNumber}
		}
	}
}

func (s *Stream) start() error {

	header := make([]byte, 2)
	header[0] = 0xA8
	header[1] = 0xA8
	// fmt.Println("After header: ", sendBuffer.Bytes())

	// Open USB interface
	inf, done, err := s.device.device.DefaultInterface()
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
		return fmt.Errorf("Full response was not recieved from device: %d != %d", n, len(recvBuffer))
	}
	// fmt.Printf("Recv buffer: %v\n", recvBuffer)

	if normalChecksum8(recvBuffer[1:]) != recvBuffer[0] {
		return ErrInvalidChecksumResponse
	} else if recvBuffer[1] != 0xA9 || recvBuffer[3] != 0x00 {
		return ErrInvalidResponseHeader
	}

	errCode := recvBuffer[2]
	if errCode != 0 {
		return fmt.Errorf("Feedback response error code (%d)", errCode)
	}
	return nil
}

func (s *Stream) stop() error {
	// s.stopCh <- struct{}{}

	header := make([]byte, 2)
	header[0] = 0xB0
	header[1] = 0xB0
	// fmt.Println("After header: ", sendBuffer.Bytes())

	// Open USB interface
	inf, done, err := s.device.device.DefaultInterface()
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

	checksumTotal8 := normalChecksum8(recvBuffer[1:])
	if checksumTotal8 != recvBuffer[0] {
		// fmt.Printf("Recv buffer: %v; checksum8=%d\n", recvBuffer, checksumTotal8)
		return ErrInvalidChecksum8Response
	} else if recvBuffer[1] != 0xB1 || recvBuffer[3] != 0x00 {
		return ErrInvalidResponseHeader
	}

	errCode := recvBuffer[2]
	if errCode != 0 && errCode != byte(52) {
		return fmt.Errorf("Feedback response error code (%d)", errCode)
	}
	return nil
}

func (s *Stream) Stop() {
	s.stopCh <- struct{}{}
}
