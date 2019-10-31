// +build js,!wasm

package quic

import (
	"fmt"
	"github.com/gopherjs/gopherjs/js"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"net"
	"time"
)

type CatalystConn struct {
	packetChan  chan []byte
	domUDP      *js.Object
	domUDPProxy *js.Object
	addr        net.Addr
	phm         *packetHandlerMap
}

func (c *CatalystConn) Write(p []byte) (int, error) {
	ui8 := make([]uint8, len(p))
	for i, b := range p {
		ui8[i] = b
	}
	c.domUDPProxy.Call("send", ui8)
	return 1, nil
}

func (c *CatalystConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	return c.Write(p)
}

func (c *CatalystConn) Read(p []byte) (n int, err error) {
	recvd := <-c.packetChan
	copied := copy(p, recvd)
	return copied, nil
}

func (c *CatalystConn) ReadFrom(p []byte) (n int, _ net.Addr, err error) {
	recvd := <-c.packetChan
	copied := copy(p, recvd)
	return copied, c.addr, nil
}

func (c *CatalystConn) SetPHM(p *packetHandlerMap) {
	c.phm = p
}

func (c *CatalystConn) Close() error {
	c.domUDP.Call("close")
	return nil
}

func (c *CatalystConn) LocalAddr() net.Addr {
	return nil
}

func (c *CatalystConn) RemoteAddr() net.Addr {
	// this is wrong, but note that Catalyst always connects to
	// the serving host and port.
	return &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 1337}
}

func (c *CatalystConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *CatalystConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *CatalystConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func newCatalystConn(addr net.Addr) *CatalystConn {
	domUDP := js.Global.Get("document").Get("realUdp")
	domUDPProxy := js.Global.Get("document").Get("realUdp")

	packetChan := make(chan []byte, 100000)
	conn := &CatalystConn{
		packetChan:  packetChan,
		domUDP:      domUDP,
		domUDPProxy: domUDPProxy,
		addr:        addr,
	}
	//go func() {
	//for {
	//data := <-packetChan
	//conn.phm.handlePacket(conn.addr, data)
	//}
	//}()
	enqueue := func(object *js.Object) {
		fmt.Println("RECEIVED")
		uint8array := js.Global.Get("Uint8Array")
		hbytes := uint8array.New(object.Get("data"))
		length := int64(hbytes.Get("byteLength").Interface().(float64))
		data := *getPacketBuffer()
		data = data[:protocol.MaxReceivePacketSize]
		bytes := hbytes.Interface().([]byte)
		copy(data, bytes)
		data = data[:length]
		fmt.Println(data)
		//go func() {
		conn.phm.handlePacket(conn.addr, data)
		//}()
		//packetChan <- data
		fmt.Println("AHH")
	}
	onclose := func() {
		fmt.Println("CLOSED")
		panic("CLOSED")
	}

	domUDP.Set("onmessage", enqueue)
	domUDP.Set("onclose", onclose)
	return conn
}
