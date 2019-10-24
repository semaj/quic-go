// +build js,wasm

package quic

import (
	"fmt"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"net"
	"syscall/js"
	"time"
)

type CatalystConn struct {
	packetChan  chan []byte
	domUDP      js.Value
	domUDPProxy js.Value
	addr        net.Addr
	phm         *packetHandlerMap
}

func (c *CatalystConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	ui8 := make([]uint8, len(p))
	for i, b := range p {
		ui8[i] = b
	}
	jsArray := js.Global().Get("Uint8Array").New(len(ui8))
	js.CopyBytesToJS(jsArray, ui8)
	c.domUDPProxy.Call("send", jsArray)
	return 1, nil
}

func (c *CatalystConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	recvd := <-c.packetChan
	copied := copy(p, recvd)
	now := time.Now()
	fmt.Println("TSLR: ", now.Sub(LastReceive))
	LastReceive = now
	return copied, c.addr, nil
}

func (c *CatalystConn) Close() error {
	c.domUDP.Call("close")
	return nil
}

func (c *CatalystConn) LocalAddr() net.Addr {
	return nil
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

func (c *CatalystConn) SetPHM(p *packetHandlerMap) {
	c.phm = p
}

func newCatalystConn(addr net.Addr) *CatalystConn {
	packetChan := make(chan []byte, 100000)
	domUDP := js.Global().Get("document").Get("realUdp")
	domUDPProxy := js.Global().Get("document").Get("realUdp")

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
	counter := 0
	t := time.Now()
	enqueue := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		//t0 := time.Now()
		counter++
		fmt.Println("CATALYSTRECEIVE ", counter)
		int8arrayWrapper := js.Global().Get("Uint8Array").New(args[0].Get("data"))
		data := *getPacketBuffer()
		data = data[:protocol.MaxReceivePacketSize]
		js.CopyBytesToGo(data, int8arrayWrapper)
		data = data[:int8arrayWrapper.Get("byteLength").Int()]
		conn.phm.handlePacket(conn.addr, data)
		//packetChan <- data
		//fmt.Println("RECEIVE ELAPSED", time.Now().Sub(t))
		t = time.Now()

		return nil
	})
	var onclose js.Func
	onclose = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		fmt.Println("CLOSED")
		panic("CLOSED")
		onclose.Release()
		return nil
	})

	domUDP.Set("onmessage", enqueue)
	domUDP.Set("onclose", onclose)
	return conn
}
