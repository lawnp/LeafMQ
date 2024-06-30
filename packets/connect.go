package packets

import "fmt"

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

func (co *ConnectOptions) Copy() *ConnectOptions {
	connectionOptions := *co
	return &connectionOptions
}

type ErrWrongProtocolName struct{}
type ErrWrongProtocolLevel struct{}
type ErrMalformedPacket struct{
	reason string
}

func (e *ErrWrongProtocolName) Error() string {
	return "Reserved bit of flag byte in CONNECT packet is not 0"
}

func (e *ErrWrongProtocolLevel) Error() string {
	return "Wrong protocol level"
}

func (e *ErrMalformedPacket) Error() string {
	return fmt.Sprintf("Malformed packet: %s", e.reason)
}

func DecodeConnect(buffer []byte) (*ConnectOptions, error) {
	protocolName, len := DecodeUTF8String(buffer)
	VersionPosition := len + 2 // 2 bytes for utf-8 len plus actual length
	protocolVersion := buffer[VersionPosition]

	switch protocolVersion {
	case 0x03:
		if protocolName != "MQIsdp" {
			return nil, &ErrWrongProtocolName{}
		}
	case 0x04, 0x05:
		if protocolName != "MQTT" {
			return nil, &ErrWrongProtocolName{}
		}
	default:
		return nil, &ErrWrongProtocolLevel{}
	}

	// in version 3.1.1 the last bit of flag byte has to be set to 0. [MQTT-3.1.2-3]
	connectFlags := buffer[VersionPosition + 1]
	if connectFlags&0x1 != 0 {
		return nil, &ErrMalformedPacket{"The last bit of the byte 'flag' needs to be 0"}
	}

	cf := DecodeConnectFlags(connectFlags)
	cf.keepalive = uint16(buffer[VersionPosition + 2])<<8 | uint16(buffer[VersionPosition + 3])

	co := DecodeConnectOptions(cf, buffer[VersionPosition + 3:])
	co.ProtocolLevel = protocolVersion
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
