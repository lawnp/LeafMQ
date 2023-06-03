package nixmq

type Client struct {
	Id string
	Username string
	Password string
	WillTopic string
	WillMessage string
	KeepAlive uint16
	CleanSession bool
}
