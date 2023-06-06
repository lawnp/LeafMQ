package packets


const UTF8BytesLength = 2

type ConnectFlags struct {
	usernameFlag bool
	passwordFlag bool
	willRetain   bool
	willQoS      byte
	willFlag     bool
	cleanSession bool
	keepalive    uint16
}

type ConnectOptions struct {
	ProtocolLevel byte
	ClientID      string
	Username      string
	Password      string
	WillTopic     string
	WillMessage   string
	WillRetain    bool
	WillQoS       byte
	CleanSession  bool
	Keepalive     uint16
}

type ErrWrongProtocolName struct{}

func (e *ErrWrongProtocolName) Error() string {
	return "Wrong message type"
}

func DecodeConnect(buffer[] byte) (*ConnectOptions, error) {
	
	if protocolName, _ := DecodeUTF8String(buffer[0:]); protocolName != "MQTT" {
		return nil, &ErrWrongProtocolName{}
	}

	flagByte := buffer[7]
	if flagByte & 0x1 != 0 {
		return nil, &ErrWrongProtocolName{}
	}

	cf := DecodeConnectFlags(flagByte)
	cf.keepalive = uint16(buffer[8])<<8 | uint16(buffer[9])

	co := DecodeConnectOptions(cf, buffer[10:])
	co.ProtocolLevel = buffer[6]
	return co, nil
}

func DecodeConnectFlags(flags byte) *ConnectFlags {
	cf := &ConnectFlags{}
	cf.usernameFlag = flags&0x80 != 0
	cf.passwordFlag = flags&0x40 != 0
	cf.willRetain = flags&0x20 != 0
	cf.willQoS = (flags & 0x18) >> 3
	cf.willFlag = flags&0x04 != 0
	cf.cleanSession = flags&0x02 != 0
	return cf
}

// DecodeConnectOptions reads the payload of the connect packet
func DecodeConnectOptions(cf *ConnectFlags, buffer []byte) *ConnectOptions {
	copts := &ConnectOptions{
		CleanSession: cf.cleanSession,
		Keepalive:    cf.keepalive,
	}

	copts.ClientID, buffer = DecodeUTF8StringInc(buffer)

	if cf.willFlag {
		copts.WillTopic, buffer = DecodeUTF8StringInc(buffer)
		copts.WillMessage, buffer = DecodeUTF8StringInc(buffer)
		copts.WillQoS = cf.willQoS
		copts.WillRetain = cf.willRetain
	}

	if cf.usernameFlag {
		copts.Username, buffer = DecodeUTF8StringInc(buffer)
	}

	if cf.passwordFlag {
		copts.Password, _ = DecodeUTF8StringInc(buffer)
	}

	if cf.passwordFlag && !cf.usernameFlag {
		// TODO: send connack packet with the proper fail return code
		panic("Password flag set but username flag not set")
	}

	return copts
}
