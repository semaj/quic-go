package main

import (
	"bytes"
	t "crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	//"syscall/js"
	//"os"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"
	//"github.com/lucas-clemente/quic-go/internal/protocol"
	//"github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {

	//f, err := os.Create("tmp/pprof")
	//if err != nil {
	//panic(err)
	//}
	roundTripper := &h2quic.RoundTripper{
		QuicConfig:      &quic.Config{},
		TLSClientConfig: &t.Config{InsecureSkipVerify: true},
	}
	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
	}

	//go func() {
	//for {
	//time.Sleep(1000 * time.Millisecond)
	//}
	//}()
	url := "https://jameslarisch.com/latency"
	payloadSizeBytes := PayloadSizeMb * 1000000.0
	for i := 0; i < NumPingPongs; i++ {
		payload := make([]byte, int64(payloadSizeBytes/float64(NumPingPongs)))
		rand.Read(payload)
		buf := bytes.NewBuffer(payload)
		fmt.Println("ABOUT TO POST")
		//js.Global().Get("window").Get("console").Call("profile", "profile1")
		t0 := time.Now()
		rsp, err := hclient.Post(url, "application/octet-stream", buf)
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		//js.Global().Get("window").Get("console").Call("profileEnd")
		fmt.Println("JUST POSTED")
		fmt.Print("LATENCY TIME ", i)
		fmt.Print(": ", t1.Sub(t0).Seconds())
		fmt.Println(" DONE")
		for {
		}

		body := &bytes.Buffer{}
		_, err = io.Copy(body, rsp.Body)
		if err != nil {
			panic(err)
		}
		rsp.Body.Close()
	}
}
