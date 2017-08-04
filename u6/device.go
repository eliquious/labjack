package u6

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// DeviceType represents the name of the device
type DeviceType string

const (

	// U6Device describes the U6 LabJack.
	U6Device = DeviceType("U6")

	// U6ProDevice describes the U6-Pro device.
	U6ProDevice = DeviceType("U6-Pro")

	// UnknownDevice describes an unknown LabJack.
	UnknownDevice = DeviceType("Unknown")
)

// DeviceDesc stores the device information.
type DeviceDesc struct {
	FirmwareVersion   string
	BootloaderVersion string
	HardwareVersion   string
	SerialNumber      int
	ProductID         int
	LocalID           int
	VersionInfo       int
	DeviceType        DeviceType
}

func (d DeviceDesc) String() string {
	return fmt.Sprintf(`U6 Device Desc: { FirmwareVersion: %s, BootloaderVersion: %s, HardwareVersion: %s, SerialNumber: %d, ProductID: %d, LocalID: %d, VersionInfo: %d, DeviceType: %s}`,
		d.FirmwareVersion,
		d.BootloaderVersion,
		d.HardwareVersion,
		d.SerialNumber,
		d.ProductID,
		d.LocalID, d.VersionInfo, d.DeviceType)
}

func parseConfigBytes(recBuffer []uint8) (DeviceDesc, error) {
	if len(recBuffer) > 38 {
		return DeviceDesc{}, errors.New("Invalid config response")
	}

	devType := UnknownDevice
	if recBuffer[37] == 4 {
		devType = U6Device
	} else if recBuffer[37] == 12 {
		devType = U6ProDevice
	}

	return DeviceDesc{
		FirmwareVersion:   fmt.Sprintf("%d.%02d", recBuffer[10], recBuffer[9]),
		BootloaderVersion: fmt.Sprintf("%d.%02d", recBuffer[12], recBuffer[11]),
		HardwareVersion:   fmt.Sprintf("%d.%02d", recBuffer[14], recBuffer[13]),
		SerialNumber:      int(binary.LittleEndian.Uint32(recBuffer[15:])),
		ProductID:         int(binary.LittleEndian.Uint16(recBuffer[19:])),
		LocalID:           int(recBuffer[21]),
		VersionInfo:       int(recBuffer[37]),
		DeviceType:        devType,
	}, nil
}
