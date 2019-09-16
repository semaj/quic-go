package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"

	_ "net/http/pprof"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"
	//"github.com/lucas-clemente/quic-go/internal/utils"
)

type binds []string

var Count int = 1000000000

func (b binds) String() string {
	return strings.Join(b, ",")
}

func (b *binds) Set(v string) error {
	*b = strings.Split(v, ",")
	return nil
}

// Size is needed by the /demo/upload handler to determine the size of the uploaded file
type Size interface {
	Size() int64
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	http.HandleFunc("/latency", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		io.Copy(w, bytes.NewBuffer([]byte("ACK")))
		Count--
		if Count == 0 {
			os.Exit(0)
		}
	})
}

func getBuildDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	return path.Dir(filename)
}

func main() {
	port := os.Args[2]
	if port == "443" {
		Count, _ = strconv.Atoi(os.Args[3])
	}
	// defer profile.Start().Stop()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// runtime.SetBlockProfileRate(1)

	//logger := utils.DefaultLogger
	//logger.SetLogLevel(utils.LogLevelDebug)
	bs := binds{}
	flag.Var(&bs, "bind", "bind to")
	tcp := flag.Bool("tcp", false, "also listen on TCP")
	flag.Parse()

	certFile := "assets/certs/fullchain.pem"
	keyFile := "assets/certs/privkey.pem"

	fs := http.FileServer(http.Dir("src/http/"))
	http.Handle("/http/", http.StripPrefix("/http/", fs))

	if len(bs) == 0 {
		bs = binds{"0.0.0.0" + port}
	}

	var wg sync.WaitGroup
	wg.Add(len(bs))
	for _, b := range bs {
		bCap := b
		go func() {
			var err error
			if *tcp {
				err = h2quic.ListenAndServe(bCap, certFile, keyFile, nil)
			} else {
				server := h2quic.Server{
					Server:     &http.Server{Addr: bCap},
					QuicConfig: &quic.Config{},
				}
				err = server.ListenAndServeTLS(certFile, keyFile)
			}
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
