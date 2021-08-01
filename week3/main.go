package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	errG, errCtx := errgroup.WithContext(context.Background())

	svr := &http.Server{
		Addr: "localhost:8080",
	}

	http.HandleFunc("/start", func(w http.ResponseWriter, req *http.Request){
		io.WriteString(w, "start\n")
	})

	shutdownCh := make(chan struct{})
	http.HandleFunc("/shutdown", func(w http.ResponseWriter, req *http.Request){
		shutdownCh <- struct{}{}
	})

	errG.Go(func() error {
		fmt.Println("start server")
		return svr.ListenAndServe()
	})

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	errG.Go(func() error {
		select {
		case <- errCtx.Done():
			fmt.Println("ctx done")
		case <- shutdownCh:
			fmt.Println("shutdown")
		case <- signalCh:
			fmt.Println("signal")
		}

		return svr.Shutdown(errCtx)
	})

	if err := errG.Wait(); err != nil {
		fmt.Printf("err happens, err:%v\n", err)
	}

	fmt.Println("finish")
}
