package packets

func DecodeUTF8String(buf []byte) (string, uint16) {
	// first two bytes are the length of the string
	length := uint16(buf[0]) << 8 | uint16(buf[1])
	return string(buf[2:length+2]), length
}

func DecodeUTF8StringInc(buf []byte) (string, []byte) {
	rez, length := DecodeUTF8String(buf)
	return rez, buf[length+2:]
}