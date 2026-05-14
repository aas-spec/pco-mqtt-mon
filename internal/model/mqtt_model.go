package model


type MQTTConfig struct {
	Host      string
	User      string
	Password  string
	Port      int
	KeepAlive int64
}
