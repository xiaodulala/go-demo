package home

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/xiaodulala/component-tools/component/json"
	"time"
)

type Client struct {
	//当前客户端连接
	clientConn *websocket.Conn
	//客户端连接地址
	address string
	//写入当前连接的消息通道
	send chan []byte
}

func NewClient(conn *websocket.Conn) (*Client, error) {
	if conn == nil {
		return nil, fmt.Errorf("websocket conn is nil")
	}
	hc := &Client{
		clientConn: conn,
		address:    conn.RemoteAddr().String(),
		send:       make(chan []byte, 1024*1024*10),
	}
	clientHub.clientRegister <- hc

	return hc, nil
}

func (c *Client) ReadMessage() {
	defer func() {
		// 断开连接,关闭连接
		if c.clientConn != nil {
			_ = c.clientConn.Close()
			clientHub.clientUnRegister <- c
		}
	}()

	for {
		msgType, msg, err := c.clientConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithFields(logrus.Fields{"address": c.address}).Warnf("readMessage err:%s", err)
			} else {
				logrus.WithFields(logrus.Fields{"address": c.address}).Infof("readMessage conn closed")
			}
			break
		}
		logrus.WithFields(logrus.Fields{"msgType": msgType, "msg": string(msg)}).Infof("read message")
		receiveData := &Msg{}
		err = json.Unmarshal(msg, receiveData)
		if err != nil {
			logrus.WithFields(logrus.Fields{"address": c.address, "error": err.Error()}).Error("msg unmarshal err")
		}
		switch receiveData.MsgType {
		case Ping:
			sendData := &Msg{MsgType: Pong}
			sendDataBytes, err := json.Marshal(sendData)
			if err != nil {
				logrus.WithFields(logrus.Fields{"msgType": sendData.MsgType, "err": err.Error()}).Warnf("json marshal err")
				continue
			}

			c.send <- sendDataBytes
		}
	}
}

func (hc *Client) WriteMessage() {
	defer func() {
		if hc.clientConn != nil {
			err := hc.clientConn.Close()
			if err != nil {
				logrus.WithFields(logrus.Fields{"address": hc.address}).Warnf("writeMessage defer in  client err:%s", err)
			}
			logrus.WithFields(logrus.Fields{"address": hc.address}).Info("writeMessage defer in  client conn closed")
		}
	}()

	for {
		select {
		case message, ok := <-hc.send:
			if !ok {
				return
			}
			//将消息管道的消息，写入当前连接
			err := hc.clientConn.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				logrus.WithFields(logrus.Fields{"address": hc.address}).Errorf("clientConn WriteMessage error:%s", err)
				return
			}
			logrus.WithFields(logrus.Fields{"address": hc.address, "msg": string(message)}).Debugf("send message success")
		}
	}
}

func (hc *Client) SendHello() {
	for {
		hc.send <- []byte("hello")
		time.Sleep(time.Second * 2)
	}
}
