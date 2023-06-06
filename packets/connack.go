package packets

import "fmt"

type Connack struct {
	FixedHeader   *FixedHeader
	SesionPresent bool
	ReturnCode    Code
}

func (c Connack) String() {
	fmt.Sprintf("FixedHeader: %+v, returnCode: %+v\n", c.FixedHeader, c.ReturnCode)
}

func NewConnack(code Code, SesionPresent bool) *Connack {
	return &Connack{
		FixedHeader: &FixedHeader{
			MessageType:     CONNACK,
			RemainingLength: 2,
		},
		SesionPresent: false,
		ReturnCode:    code,
	}
}

func (cack *Connack) Encode() []byte {
	buffer := make([]byte, 4)
	buffer[0] = 0x2 << 4
	buffer[1] = 0x2

	// ugly but go is kinda stupid when it comes to booleans
	if cack.SesionPresent {
		buffer[2] = 0x1
	} else {
		buffer[2] = 0x0
	}

	buffer[3] = cack.ReturnCode.Code
	return buffer
}
