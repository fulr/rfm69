package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/fulr/rfm69"
	"github.com/fulr/rfm69/payload"
)

// Configuration defines the config options and file structure
type Configuration struct {
	EncryptionKey string
	NodeID        byte
	NetworkID     byte
	IsRfm69Hw     bool
	MqttBroker    string
	MqttClientID  string
}

var defautlPubHandler = func(client *MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func actorHandler(tx chan *rfm69.Data) func(client *MQTT.Client, msg MQTT.Message) {
	return func(client *MQTT.Client, msg MQTT.Message) {
		command := string(msg.Payload())
		log.Println(msg.Topic(), command)
		on := byte(0)
		if command == "ON" {
			on = 1
		}
		parts := strings.Split(msg.Topic(), "/")
		node, err := strconv.Atoi(parts[2])
		if err != nil {
			log.Println(err)
			return
		}
		pin, err := strconv.Atoi(parts[3])
		if err != nil {
			log.Println(err)
			return
		}
		buf := bytes.Buffer{}
		binary.Write(&buf, binary.LittleEndian, payload.Payload{Type: 2, Uptime: 1})
		binary.Write(&buf, binary.LittleEndian, payload.Payload2{Pin: byte(pin), State: on})
		tx <- &rfm69.Data{
			ToAddress:  byte(node),
			Data:       buf.Bytes(),
			RequestAck: true,
		}
	}
}

func readConfig() (*Configuration, error) {
	file, err := os.Open("conf.json")
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err = decoder.Decode(&config)
	file.Close()
	return &config, err
}

func main() {
	log.Print("Reading config")
	config, err := readConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(config)
	opts := MQTT.NewClientOptions().AddBroker(config.MqttBroker).SetClientID(config.MqttClientID)
	opts.SetDefaultPublishHandler(defautlPubHandler)
	opts.SetCleanSession(true)
	c := MQTT.NewClient(opts)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	rfm, err := rfm69.NewDevice(config.NodeID, config.NetworkID, config.IsRfm69Hw)
	if err != nil {
		log.Fatal(err)
	}
	defer rfm.Close()
	err = rfm.Encrypt([]byte(config.EncryptionKey))
	if err != nil {
		log.Fatal(err)
	}
	rx, tx, quit := rfm.Loop()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, os.Kill)

	token = c.Subscribe("/actor/#", 0, actorHandler(tx))
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	defer c.Unsubscribe("/actor/#")

	for {
		select {
		case data := <-rx:
			if data.ToAddress != config.NodeID {
				break
			}
			log.Println("got data from", data.FromAddress, ", RSSI", data.Rssi)
			if data.ToAddress != 255 && data.RequestAck {
				tx <- data.ToAck()
			}
			topic := fmt.Sprintf("/sensor/%d/", data.FromAddress)
			pubToken := c.Publish(topic+"rssi", 0, false, fmt.Sprintf("%d", data.Rssi))
			pubToken.Wait()
			if len(data.Data) > 5 {
				var p payload.Payload
				buf := bytes.NewReader(data.Data)
				binary.Read(buf, binary.LittleEndian, &p)
				log.Println("payload", p)
				switch p.Type {
				case 1:
					var p1 payload.Payload1
					binary.Read(buf, binary.LittleEndian, &p1)
					log.Println("payload1", p1)
					pubToken = c.Publish(topic+"temp", 0, false, fmt.Sprintf("%f", p1.Temperature))
					pubToken.Wait()
					pubToken = c.Publish(topic+"hum", 0, false, fmt.Sprintf("%f", p1.Humidity))
					pubToken.Wait()
					pubToken = c.Publish(topic+"bat", 0, false, fmt.Sprintf("%f", p1.VBat))
					pubToken.Wait()
				default:
					log.Println("unknown payload")
				}
			}
		case <-sigint:
			quit <- true
			<-quit
			c.Disconnect(250)
			return
		}
	}
}
