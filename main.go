package main

import (
	"flag"
	"net/http"
	"net/rpc"
	"net/rpc/jsonrpc"

	"github.com/golang/glog"
)

type rwc struct {
	res http.ResponseWriter
	req *http.Request
}

func (r rwc) Close() error {
	return r.req.Body.Close()
}

func (r rwc) Read(p []byte) (int, error) {
	return r.req.Body.Read(p)
}

func (r rwc) Write(p []byte) (int, error) {
	return r.res.Write(p)
}

func main() {
	defer glog.Flush()

	flag.Parse()

	glog.Info("starting listener")

	service := NewService()
	rpc.Register(service)
	http.HandleFunc("/jsonrpc", func(res http.ResponseWriter, req *http.Request) {
		glog.Infof("rpc")
		res.Header().Set("Content-Type", "application/json")
		rpc.ServeRequest(jsonrpc.NewServerCodec(rwc{res, req}))
	})

	err := http.ListenAndServe("127.0.0.1:9999", nil)
	if err != nil {
		glog.Fatalln(err)
	}
}
