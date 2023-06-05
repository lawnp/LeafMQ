package packets

type Code struct {
	Code byte
	Reason string
}

var (
	ACCEPTED = Code{0x00, "Connection Accepted"}
	UNACCEPTABLE_PROTOCOL_VERSION = Code{0x01, "Connection Refused, unacceptable protocol version"}
	IDENTIFIER_REJECTED = Code{0x02, "Connection Refused, identifier rejected"}
	SERVER_UNAVAILABLE = Code{0x03, "Connection Refused, server unavailable"}
	BAD_USERNAME_OR_PASSWORD = Code{0x04, "Connection Refused, bad user name or password"}
	NOT_AUTHORIZED = Code{0x05, "Connection Refused, not authorized"}
)