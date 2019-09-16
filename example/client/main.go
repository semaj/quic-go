package main

import (
	"bytes"
	t "crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"syscall/js"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"
	//"github.com/lucas-clemente/quic-go/internal/protocol"
	//"github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {

	//logger := utils.DefaultLogger

	//logger.SetLogLevel(utils.LogLevelDebug)
	//logger.SetLogLevel(utils.LogLevelDebug)
	//logger.SetLogTimeFormat("")

	//versions := protocol.SupportedVersions
	payloadSizeMb := js.Global().Get("payloadSizeMb").Int()
	numPingPongs := js.Global().Get("numPingPongs").Int()

	roundTripper := &h2quic.RoundTripper{
		QuicConfig:      &quic.Config{},
		TLSClientConfig: &t.Config{InsecureSkipVerify: true},
	}
	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
	}

	url := "https://jameslarisch.com/latency"
	payloadSizeBytes := payloadSizeMb * 1000000
	for i := 0; i < numPingPongs; i++ {
		payload := make([]byte, payloadSizeBytes/numPingPongs)
		rand.Read(payload)
		t0 := time.Now()
		rsp, err := hclient.Post(url, "application/octet-stream", bytes.NewBuffer(payload))
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		fmt.Print("LATENCY TIME", i)
		fmt.Print(":", t1.Sub(t0).Seconds())
		fmt.Println(" DONE")

		body := &bytes.Buffer{}
		_, err = io.Copy(body, rsp.Body)
		if err != nil {
			panic(err)
		}
		rsp.Body.Close()
	}
}
