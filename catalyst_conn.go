// +build js,wasm

package quic

import (
	//b64 "encoding/base64"
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
	//ui8 := make([]uint8, len(p))
	//for i, b := range p {
	//ui8[i] = b
	//}
	jsArray := js.Global().Get("Uint8Array").New(len(p))
	js.CopyBytesToJS(jsArray, p)
	//fmt.Println("ENCODE SIZE", len(p))
	//fmt.Println("PACKET SEND START")
	//sEnc := b64.StdEncoding.EncodeToString(p)
	//fmt.Println("PACKET SEND STOP")
	c.domUDPProxy.Call("send", jsArray)
	return 1, nil
}

func (c *CatalystConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	//time.Sleep(1 * time.Millisecond)
	//s0 := time.Now()
	recvd := <-c.packetChan
	//fmt.Println("Elapsed ", time.Now().Sub(s0))
	copied := copy(p, recvd)
	now := time.Now()
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
		go func() {
			counter++
			//fmt.Println("RECEIVEDCOUNTER", counter)
			//data, err := b64.StdEncoding.DecodeString(args[0].String())
			//if err != nil {
			//panic(err)
			//}
			int8arrayWrapper := js.Global().Get("Uint8Array").New(args[0].Get("data"))
			data := *getPacketBuffer()
			data = data[:protocol.MaxReceivePacketSize]
			js.CopyBytesToGo(data, int8arrayWrapper)
			byteLength := int8arrayWrapper.Get("byteLength").Int()
			CatLog("ByteLength", byteLength)
			data = data[:byteLength]
			//conn.phm.handlePacket(conn.addr, data)
			packetChan <- data
			//fmt.Println("RECEIVE ELAPSED", time.Now().Sub(t))
			t = time.Now()
		}()

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
	domUDPProxy.Set("onclose", onclose)
	return conn
}
