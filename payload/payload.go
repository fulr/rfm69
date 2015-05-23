package payload

// Payload is the first part
type Payload struct {
	Type   int16  // sensor type (2, 3, 4, 5)
	Uptime uint32 // uptime in ms
}

// Payload1 is for DHT22-nodes
type Payload1 struct {
	Temperature float32 // Temp
	Humidity    float32 // Humidity
	VBat        float32 // V Battery
}

// Payload2 is for switches
type Payload2 struct {
	Pin   byte // Pin number
	State byte // 1=On 0=Off
}
