package modbus

// Cyclical Redundancy Checking.
var (
	crtTable []uint16
)

// CRC16 Calculate Cyclical Redundancy Checking.
func CRC16(bs []byte) uint16 {
	val := uint16(0xFFFF)
	for _, v := range bs {
		val = (val >> 8) ^ crtTable[(val^uint16(v))&0x00FF]
	}
	return val
}

// init 初始化表.
func init() {
	crcPoly16 := uint16(0xa001)
	crtTable = make([]uint16, 256)

	for i := uint16(0); i < 256; i++ {
		crc := uint16(0)
		b := i

		for j := uint16(0); j < 8; j++ {
			if ((crc ^ b) & 0x0001) > 0 {
				crc = (crc >> 1) ^ crcPoly16
			} else {
				crc >>= 1
			}
			b >>= 1
		}
		crtTable[i] = crc
	}
}
