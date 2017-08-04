package u6

import (
	"errors"
	"fmt"
)

// ErrInvalidContext is returned if the context is nil
var ErrInvalidContext = errors.New("Invalid USB context")

// ErrEndpointSendError is returned when data could not be sent or not all the data was sent.
var ErrEndpointSendError = errors.New("Failed to send data to device")

// ErrEndpointRecvError is returned when data could not be read or not all the data was received.
var ErrEndpointRecvError = errors.New("Failed to receive data from device")

// ErrInvalidChecksumInput occurs when the checksum input is too short
var ErrInvalidChecksumInput = errors.New("Checksum could not be calculated; input too short")

// ErrInvalidChecksumResponse is returned if the U6 detected a bad checksum
var ErrInvalidChecksumResponse = errors.New("The U6 detected a bad checksum. Double check your checksum calculations and try again")

// ErrInvalidChecksum is returned if the checksum is invalid
var ErrInvalidChecksum = errors.New("Invalid checksum")

// ErrInvalidResponseHeader is returned if the response header is not valid.
var ErrInvalidResponseHeader = errors.New("Invalid response header")

// ErrResponseTooShort is returned if the response cannot be validated due to short length.
var ErrResponseTooShort = errors.New("Response data is too short")

// ErrLibUSB returns when there's a low-level error in gousb.
type ErrLibUSB struct {
	message string
	err     error
}

func (e ErrLibUSB) Error() string {
	return fmt.Sprintf("%s: %v", e.message, e.err)
}

// ErrLabJackErrorCode returns when there's an error response from the U6.
type ErrLabJackErrorCode struct {
	code int
}

func (e ErrLabJackErrorCode) Error() string {
	return fmt.Sprintf("LabJack error code: %v", e.code)
}
