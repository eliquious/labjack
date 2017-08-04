package u6

// normalChecksum returns an 8-bit checksum
func normalChecksum(bytes []uint8) uint8 {
	return normalChecksum8(bytes)
}

// normalChecksum8 calculates the 8-bit checksum
func normalChecksum8(bytes []uint8) uint8 {
	var a, bb uint16

	//Sums bytes 1 to n-1 unsigned to a 2 byte value. Sums quotient and
	//remainder of 256 division.  Again, sums quotient and remainder of
	//256 division.
	for _, b := range bytes {
		a += uint16(b)
	}

	bb = a / 256
	a = (a - 256*bb) + bb
	bb = a / 256
	return uint8((a - 256*bb) + bb)
}

// extendedChecksum in-lines the 16-bit checksum in the slice
func extendedChecksum(bytes []uint8) error {
	a, err := extendedChecksum16(bytes)
	if err != nil {
		return err
	}

	b, err := extendedChecksum8(bytes)
	if err != nil {
		return err
	}

	bytes[4] = uint8(a & 0xff)
	bytes[5] = uint8((a / 256) & 0xff)
	bytes[0] = b
	return nil
}

// extendedChecksum16 returns the 16-bit checksum
func extendedChecksum16(bytes []uint8) (uint16, error) {
	if len(bytes) < 7 {
		return 0, ErrInvalidChecksumInput
	}
	var a uint16

	//Sums bytes 6 to n-1 to a unsigned 2 byte value
	for i := 6; i < len(bytes); i++ {
		a += uint16(bytes[i])
	}
	return a, nil
}

// extendedChecksum8 returns the 8-bit extended checksum
func extendedChecksum8(bytes []uint8) (uint8, error) {
	if len(bytes) < 6 {
		return 0, ErrInvalidChecksumInput
	}
	var a, bb int

	//Sums bytes 1 to 5. Sums quotient and remainder of 256 division. Again, sums
	//quotient and remainder of 256 division.
	for i := 1; i < 6; i++ {
		a += int(uint16(bytes[i]))
	}

	bb = a / 256
	a = (a - 256*bb) + bb
	bb = a / 256

	return uint8((a - 256*bb) + bb), nil
}

func setChecksum(bytes []uint8) error {
	if len(bytes) < 6 {
		return ErrInvalidChecksumInput
	}

	a := bytes[1]
	a = (a & 0x78) >> 3
	if a == 15 {
		setChecksum16(bytes)
		setChecksum8(bytes, 6)
	} else {
		setChecksum8(bytes, len(bytes))
	}
	return nil
}

func setChecksum16(bytes []uint8) {
	var total uint8
	for i := 6; i < len(bytes); i++ {
		total += bytes[i] & 0xFF
	}
	bytes[4] = total & 0xFF
	bytes[5] = (total >> 8) & 0xFF
}

func setChecksum8(bytes []uint8, num int) {
	var total uint8
	for i := 1; i < num; i++ {
		total += bytes[i] & 0xFF
	}
	bytes[0] = total&0xFF + ((total >> 8) & 0xFF)
	bytes[0] = bytes[0]&0xFF + ((bytes[0] >> 8) & 0xFF) + 1
}

func uint8ArrayToFloat64(buffer []uint8, startIndex int) float64 {
	var resultDec uint32
	var resultWh int32
	resultDec = (uint32(buffer[startIndex]) |
		(uint32(buffer[startIndex+1]) << 8) |
		(uint32(buffer[startIndex+2]) << 16) |
		(uint32(buffer[startIndex+3]) << 24))
	resultWh = (int32(buffer[startIndex+4]) |
		(int32(buffer[startIndex+5]) << 8) |
		(int32(buffer[startIndex+6]) << 16) |
		(int32(buffer[startIndex+7]) << 24))
	return float64(int(resultWh)) + float64(resultDec)/4294967296.0
}
