package packets

import (
	"fmt"
)

const UTF8BytesLength = 2

type ConnectFlags struct {
	usernameFlag bool
	passwordFlag bool
	willRetain bool
	willQoS byte
	willFlag bool
	cleanSession bool
	keepalive uint16
}

type ConnectOptions struct {
	ConnectFlags *ConnectFlags
	clientID string
	willTopic string
	willMessage string
	username string
	password string
}

func ParseConnect(packet []byte) *ConnectOptions {
	fh := DecodeFixedHeader(packet)

	if fh.messageType != CONNECT {
		// TODO: send connack packet with the proper fail return code
		panic("Wrong message type")
	}

	if protocolName, _ := DecodeUTF8String(packet[2:]); protocolName != "MQTT" {
		// TODO: disconnect client
		fmt.Println(protocolName)
		panic("Wrong protocol name")
	}

	protocolLevel := packet[8]

	if protocolLevel != 4 {
		// TODO send connack packet with the proper fail return code
		fmt.Println(protocolLevel)
		panic("Wrong protocol level")
	}

	cf := parseFlags(packet[9])
	cf.keepalive = uint16(packet[10]) << 8 | uint16(packet[11])
	return parseConnectOptions(cf, packet[12:])

}

func parseFlags(flags byte) *ConnectFlags {
	cf := &ConnectFlags{}
	cf.usernameFlag = flags & 0x80 != 0
	cf.passwordFlag = flags & 0x40 != 0
	cf.willRetain = flags & 0x20 != 0
	cf.willQoS = (flags & 0x18) >> 3
	cf.willFlag = flags & 0x04 != 0
	cf.cleanSession = flags & 0x02 != 0
	return cf
}

func parseConnectOptions(cf *ConnectFlags, packet []byte) *ConnectOptions {
	copts := &ConnectOptions{
		ConnectFlags: cf,
	}

	copts.clientID, packet = DecodeUTF8StringInc(packet)

	if cf.willFlag {
		copts.willTopic, packet = DecodeUTF8StringInc(packet)
		copts.willMessage, packet = DecodeUTF8StringInc(packet)
	}

	if cf.usernameFlag {
		copts.username, packet = DecodeUTF8StringInc(packet)
	}

	if cf.passwordFlag {
		copts.password, _ = DecodeUTF8StringInc(packet)
	}

	if cf.passwordFlag && !cf.usernameFlag {
		// TODO: send connack packet with the proper fail return code
		panic("Password flag set but username flag not set")
	}

	return copts
}