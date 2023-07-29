package packets

type Connack struct {
	FixedHeader   *FixedHeader
	SessionPresent bool
	ReturnCode    Code
}

func NewConnack(code Code, sessionPresent bool) *Connack {
	return &Connack{
		FixedHeader: &FixedHeader{
			MessageType:     CONNACK,
			RemainingLength: 2,
		},
		SessionPresent: sessionPresent,
		ReturnCode:    code,
	}
}

func (cack *Connack) Encode() []byte {
	buffer := make([]byte, 4)
	buffer[0] = 0x2 << 4
	buffer[1] = 0x2

	// ugly but go is kinda stupid when it comes to booleans
	if cack.SessionPresent {
		buffer[2] = 0x1
	} else {
		buffer[2] = 0x0
	}

	buffer[3] = cack.ReturnCode.Code
	return buffer
}
