/*
Package labjack re-implements the LabJack USB communication and APIs in Go. Vendor and Product IDs come from the LabJack exodriver code base.
*/
package labjack

import (
	"github.com/google/gousb"
)

// LabJackVendorID is the ID for the LabJack company.
const LabJackVendorID gousb.ID = gousb.ID(0x0cd5)

// U6ProductID is the ID for the U6 / U6 Pro devices
const U6ProductID gousb.ID = gousb.ID(0x0006)

// U6 pipes to read/write through
const (
	U6PipeOutEP1 int = 1
	U6PipeInEP2  int = 0x82
	U6PipeInEP3  int = 0x83 //Stream Endpoint
)
