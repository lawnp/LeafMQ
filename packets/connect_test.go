package packets

import (
	"testing"
)

func TestDecodeConnect(t *testing.T) {
    packetV3 := append([]byte{0x00, 0x06}, []byte("MQIsdp")...)
    packetV5 := append([]byte{0x00, 0x04}, []byte("mqtt")...)

    // versions
    packetV3 = append(packetV3, 0x03)
    packetV4 := append(packetV5, 0x04)
    packetV5 = append(packetV5, 0x05)


    rest := []byte{0x00, 0x00, 0x00} // flags and keepAlive

    packetV3 = append(packetV3, rest...)
    packetV5 = append(packetV5, rest...)
    packetV5 = append(packetV5, rest...)

    connectionOptionsV3, err := DecodeConnect(packetV3)
    if err != nil {
        t.Errorf(err.Error())
    }

    if connectionOptionsV3.ProtocolLevel != 3 {
        t.Errorf("Expected connect packet to have version 3 but got: %b", connectionOptionsV3.ProtocolLevel)
    }

    connectionOptionsV4, err := DecodeConnect(packetV4)
    if err != nil {
        t.Errorf(err.Error())
    }

    if connectionOptionsV4.ProtocolLevel != 4 {
        t.Errorf("Expected connect packet to have version 4 but got: %b", connectionOptionsV3.ProtocolLevel)
    }


    connectionOptionsV5, err := DecodeConnect(packetV5)
    if err != nil {
        t.Errorf(err.Error())
    }

    if connectionOptionsV5.ProtocolLevel != 5 {
        t.Errorf("Expected connect packet to have version 5 but got: %b", connectionOptionsV3.ProtocolLevel)
    }
}
