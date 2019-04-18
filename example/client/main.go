package main

import (
  "fmt"
  "net/http"
  "crypto/tls"
  "strconv"
  "time"
  //"io/ioutil"

  quic "github.com/lucas-clemente/quic-go"
  "github.com/lucas-clemente/quic-go/http3"
  "github.com/lucas-clemente/quic-go/internal/protocol"
)

func main() {
  versions := protocol.SupportedVersions

  roundTripper := &http3.RoundTripper{
    QuicConfig: &quic.Config{Versions: versions},
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  hclient := &http.Client{
    Transport: roundTripper,
  }
  p := true
  start := 0
  stop := 1000
  for i := start; i < stop; i++ {
    t0 := time.Now()
    //url := "https://stalepopcorn.club/static/files/file1.html"
    url := "https://stalepopcorn.club/static/files/file"+strconv.Itoa(i)+".html"
    //url := "https://stalepopcorn.club/random"
    fmt.Println(url)
    _, err := hclient.Get(url)
    //body, err := ioutil.ReadAll(resp.Body)
    //fmt.Println(len(body))
    //resp.Body.Close()
    if err != nil {
      fmt.Println(err)
      panic("failed")
    } else {
      t1 := time.Now()
      elapsed := t1.Sub(t0)
      if (p) {
        fmt.Println(elapsed)
      }
      p = true
    }
  }
  fmt.Println("Done!")
  roundTripper.Close()
}
