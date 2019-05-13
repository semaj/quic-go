package main

import (
	"flag"
    "io"
	"log"
	"net/http"
	"path"
	"runtime"
	"strings"
    "math/rand"
	"sync"

	_ "net/http/pprof"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/h2quic"
    "github.com/lucas-clemente/quic-go/internal/utils"
)

type binds []string

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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", "no-cache");
        io.Copy(w, r.Body);
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
	// defer profile.Start().Stop()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// runtime.SetBlockProfileRate(1)

    logger := utils.DefaultLogger
    logger.SetLogLevel(utils.LogLevelDebug)
	bs := binds{}
	flag.Var(&bs, "bind", "bind to")
	certPath := flag.String("certpath", getBuildDir(), "certificate directory")
	tcp := flag.Bool("tcp", false, "also listen on TCP")
	flag.Parse()

	certFile := *certPath + "/fullchain.pem"
	keyFile := *certPath + "/privkey.pem"

    fs := http.FileServer(http.Dir("/home/james/catalyst-benchmarks/assets/"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	if len(bs) == 0 {
		bs = binds{"localhost:443"}
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
