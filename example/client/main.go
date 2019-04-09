package main

import (
  "bytes"
  "flag"
  "io"
  "net/http"
  t "crypto/tls"

  quic "github.com/lucas-clemente/quic-go"
  "github.com/lucas-clemente/quic-go/h2quic"
  "github.com/lucas-clemente/quic-go/internal/protocol"
  "github.com/lucas-clemente/quic-go/internal/utils"
)

func main() {
  verbose := flag.Bool("v", false, "verbose")
  tls := flag.Bool("tls", false, "activate support for IETF QUIC (work in progress)")
  quiet := flag.Bool("q", false, "don't print the data")
  flag.Parse()
  urls := flag.Args()

  logger := utils.DefaultLogger

  if *verbose {
    logger.SetLogLevel(utils.LogLevelDebug)
  } else {
    logger.SetLogLevel(utils.LogLevelInfo)
  }
  logger.SetLogTimeFormat("")

  versions := protocol.SupportedVersions
  if *tls {
    versions = append([]protocol.VersionNumber{protocol.VersionTLS}, versions...)
  }

  roundTripper := &h2quic.RoundTripper{
    QuicConfig: &quic.Config{Versions: versions},
    TLSClientConfig: &t.Config{InsecureSkipVerify: true},
  }
  defer roundTripper.Close()
  hclient := &http.Client{
    Transport: roundTripper,
  }

  for i := 0; i < 1000; i++ {
    rsp, err := hclient.Post(urls[0], "application/octet-stream", bytes.NewBuffer([]byte("HELLO")))
    if err != nil {
      panic(err)
    }
    logger.Infof("%d Got response", i)

    body := &bytes.Buffer{}
    _, err = io.Copy(body, rsp.Body)
    if err != nil {
      panic(err)
    }
    if *quiet {
      logger.Infof("Request Body: %d bytes", body.Len())
    } else {
      logger.Infof("Request Body:")
      logger.Infof("%s", body.Bytes())
    }
  }
}
