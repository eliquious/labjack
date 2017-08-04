package u6

import "testing"

func TestExtendedChecksum(t *testing.T) {
	command := make([]byte, 8)
	command[1] = 0xF8
	command[2] = 0x01
	command[3] = 0x2D
	command[6] = 0x00
	command[7] = 0x00

	err := setChecksum(command)
	if err != nil {
		t.Fatalf("Checksum error")
	}
	t.Logf("After: %v", command)

	command[7] = 0x02
	err = setChecksum(command)
	if err != nil {
		t.Fatalf("Checksum error")
	}
	t.Logf("After: %v", command)
}

func TestUint8ArrayToFloat64(t *testing.T) {
	data := []byte{0, 0, 148, 38, 54, 131, 0, 0}
	if uint8ArrayToFloat64(data, 0) != 33590.15069580078 {
		t.Logf("Not Equal: %v != %v", uint8ArrayToFloat64(data, 0), 33590.15069580078)
	}
	t.Logf("Equal: %v == %v", uint8ArrayToFloat64(data, 0), 33590.15069580078)

	data = []byte{190, 139, 221, 228, 255, 255, 255, 255}
	if uint8ArrayToFloat64(data, 0) != -0.10599447833374143 {
		t.Logf("Not Equal: %v != %v", uint8ArrayToFloat64(data, 0), -0.10599447833374143)
	}
	t.Logf("Equal: %v == %v", uint8ArrayToFloat64(data, 0), -0.10599447833374143)
}
