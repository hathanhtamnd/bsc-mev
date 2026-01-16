package mev

import (
	"encoding/json"
	"net"
	"sync"
)

var (
	once    sync.Once
	udpConn *net.UDPConn
)

const javaUDPAddr = "127.0.0.1:8999"

func initUDP() {
	raddr, err := net.ResolveUDPAddr("udp", javaUDPAddr)
	if err != nil {
		return
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return
	}
	udpConn = conn
}

func send(ev *TxEvent) {
	once.Do(initUDP)
	if udpConn == nil {
		return
	}
	b, _ := json.Marshal(ev)
	_, _ = udpConn.Write(b)
}
