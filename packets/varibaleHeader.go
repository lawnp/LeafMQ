package packets

type VariableHeader struct {}

func DecodeVariableHeader(buffer []byte) *VariableHeader {
	return &VariableHeader{}
}