package main

import (
	"bytes"
	t "crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"
	//"github.com/lucas-clemente/quic-go/internal/protocol"
	//"github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {
	urls := [1]string{"https://jameslarisch.com/"}

	//logger := utils.DefaultLogger

	//logger.SetLogLevel(utils.LogLevelDebug)
	//logger.SetLogLevel(utils.LogLevelDebug)
	//logger.SetLogTimeFormat("")

	//versions := protocol.SupportedVersions

	roundTripper := &h2quic.RoundTripper{
		QuicConfig:      &quic.Config{},
		TLSClientConfig: &t.Config{InsecureSkipVerify: true},
	}
	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
	}

	payload := make([]byte, 16384)
	latencies := make([]time.Duration, 1000)
	for i := 0; i < 1000; i++ {
		rand.Read(payload)
		t0 := time.Now()
		rsp, err := hclient.Post(urls[0], "application/octet-stream", bytes.NewBuffer(payload))
		if err != nil {
			panic(err)
		}
		t1 := time.Now()
		latencies[i] = t1.Sub(t0)

		body := &bytes.Buffer{}
		_, err = io.Copy(body, rsp.Body)
		if err != nil {
			panic(err)
		}
		rsp.Body.Close()
	}
	fmt.Println(latencies)
	sum := int64(0)
	for i := 0; i < len(latencies); i++ {
		sum += latencies[i].Nanoseconds() / 1000000
	}
	fmt.Println(sum / int64(len(latencies)))
}
