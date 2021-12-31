package home

import (
	"github.com/sirupsen/logrus"
	"sync"
)

type ClientHub struct {
	clients          map[string]*Client
	mutex            sync.Mutex
	clientRegister   chan *Client
	clientUnRegister chan *Client
	broadcast        chan []byte
}

var clientHub *ClientHub

func InitClientHub() {
	clientHub = newClientHub()
	go clientHub.run()
}

func newClientHub() *ClientHub {
	return &ClientHub{
		clients:          make(map[string]*Client),
		mutex:            sync.Mutex{},
		clientRegister:   make(chan *Client),
		clientUnRegister: make(chan *Client),
		broadcast:        make(chan []byte, 100),
	}
}

//func destroyHomeClientHub() {
//	for _, client := range clientHub.clients {
//		if client != nil {
//			_ = client.clientConn.Close()
//			close(client.send)
//		}
//	}
//	close(clientHub.clientRegister)
//	close(clientHub.clientUnRegister)
//	clientHub = nil
//}

func (hc *ClientHub) run() {
	for {
		select {
		case homeClient := <-hc.clientRegister:
			if homeClient == nil {
				continue
			}
			hc.mutex.Lock()
			if _, ok := hc.clients[homeClient.address]; !ok {
				hc.clients[homeClient.address] = homeClient
			}
			hc.mutex.Unlock()
			logrus.WithFields(logrus.Fields{"address": homeClient.address}).Infof("client register success")
		case homeClient := <-hc.clientUnRegister:
			if homeClient != nil {
				continue
			}
			hc.mutex.Lock()
			if _, ok := hc.clients[homeClient.address]; ok {
				delete(hc.clients, homeClient.address)
			}
			hc.mutex.Unlock()
			logrus.WithFields(logrus.Fields{"address": homeClient.address}).Infof("client unRegister success")
		case msg := <-hc.broadcast:
			for address, client := range hc.clients {
				select {
				case client.send <- msg:
				default:
					logrus.WithFields(logrus.Fields{"address": address}).Warnf("client send channle is full")
				}
			}
		}
	}
}
