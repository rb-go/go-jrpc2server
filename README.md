# fasthttp_json_rpc2 (early beta)
[Website](https://www.riftbit.com) | [Contributing](https://www.riftbit.com/How-to-Contribute)

[![license](https://img.shields.io/github/license/riftbit/fasthttp_json_rpc2.svg)](LICENSE)
[![GoDoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/riftbit/fasthttp_json_rpc2)
[![Coverage Status](https://coveralls.io/repos/github/riftbit/fasthttp_json_rpc2/badge.svg?branch=master)](https://coveralls.io/github/riftbit/fasthttp_json_rpc2?branch=master)
[![Build Status](https://travis-ci.org/riftbit/fasthttp_json_rpc2.svg?branch=master)](https://travis-ci.org/riftbit/fasthttp_json_rpc2)
[![Go Report Card](https://goreportcard.com/badge/github.com/riftbit/fasthttp_json_rpc2)](https://goreportcard.com/report/github.com/riftbit/fasthttp_json_rpc2)

## System requirements 
- github.com/valyala/fasthttp
- tested on golang 1.10

## Examples

### Benchmark results
Tested on work PC (4 cores, 16gb memory, windows 10)
Used memory on all benchmark time: 7.2 MB
Used CPU on all benchmark time: 12%
Load soft used on same PC: SuperBenchmark (sb)

![Benchmark Results](rps_results.png?raw=true "Benchmark Results")

### Without routing

```golang
package main

import (
	"github.com/valyala/fasthttp"
	"github.com/riftbit/fasthttp_json_rpc2"
	"log"
	"runtime"
	"runtime/debug"
)

//Log area
type DemoAPI struct{}

//Add Method to add Log
func (h *DemoAPI) Test(ctx *fasthttp.RequestCtx, args *struct{ID string}, reply *struct{LogID string}) error {
	//log.Println(args)
	reply.LogID = args.ID
	return nil
}

func init() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU*10)
	debug.SetMaxThreads(250)
}

func main() {
	api := jsonrpc2.NewServer()
	err := api.RegisterService(new(DemoAPI), "demo")

	if err != nil {
		log.Fatalln(err)
	}

	fasthttp.ListenAndServe(":8081", api.APIHandler)
}
```


### With minimal routing

```golang
package main

import (
	"github.com/valyala/fasthttp"
	"github.com/riftbit/fasthttp_json_rpc2"
	"log"
	"runtime"
	"runtime/debug"
)

//Log area
type DemoAPI struct{}

//Add Method to add Log
func (h *DemoAPI) Test(ctx *fasthttp.RequestCtx, args *struct{ID string}, reply *struct{LogID string}) error {
	//log.Println(args)
	reply.LogID = args.ID
	return nil
}

func init() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU*10)
	debug.SetMaxThreads(250)
}

func main() {
	api := jsonrpc2.NewServer()
	err := api.RegisterService(new(DemoAPI), "demo")

	if err != nil {
		log.Fatalln(err)
	}

	reqHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/api":
			api.APIHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	fasthttp.ListenAndServe(":8081", reqHandler)
}
```

### With advanced routing and middlewares (not tested, will be updated)

```golang
package main

import (
	"github.com/valyala/fasthttp"
	"github.com/riftbit/fasthttp_json_rpc2"
	"log"
	"runtime"
	"runtime/debug"
)

//Log area
type DemoAPI struct{}

//Add Method to add Log
func (h *DemoAPI) Test(ctx *fasthttp.RequestCtx, args *struct{ID string}, reply *struct{LogID string}) error {
	//log.Println(args)
	reply.LogID = args.ID
	return nil
}

func init() {
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU*10)
	debug.SetMaxThreads(250)
}

func main() {
	api := jsonrpc2.NewServer()
	err := api.RegisterService(new(DemoAPI), "demo")
	if err != nil {
		log.Fatalln(err)
	}

	router := routing.New()

    router.Post("/api", StartHandler, api.APIHandler, FinishHandler)
    router.Post("/<node>/*", OptionsHandler)
    router.Get("*", OptionsHandler)
    router.Options("*", OptionsHandler)


	server := fasthttp.Server{
		Name:    "AWESOME SERVER BY RIFTBIT.COM",
		Handler: router.HandleRequest,
	}

	Logger.Println("Started fasthttp server on port", config.System.ListenOn)
	Logger.Fatal(server.ListenAndServe(config.System.ListenOn))
}


func OptionsHandler(ctx *routing.Context) error {
	ctx.Response.SetBody([]byte(`{"error": "this method not allowed"}`))
	ctx.Response.SetStatusCode(405)
	setServerHeaders(ctx)
	ctx.Abort()
	return nil
}

func StartHandler(ctx *routing.Context) error {
	nodeData, ok := NodesList[ctx.Param("node")]
	if ok != true {
		ctx.Response.SetBody([]byte(`{"error": "node not found"}`))
		ctx.Response.SetStatusCode(404)
		setServerHeaders(ctx)
		ctx.Abort()
		return nil
	}
	ctx.Set("TimeStarted", time.Now())
	ctx.Set("NodeData", nodeData)
	ctx.Set("UrlParams", strings.TrimPrefix(string(ctx.Path()), "/"+string(ctx.Param("node"))))
	return nil
}

func FinishHandler(ctx *routing.Context) error {
	startedAt := ctx.Get("TimeStarted").(time.Time)
	NodeData := ctx.Get("NodeData").(nodeElement)
	timeFinished := time.Since(startedAt)
	urlPath := ctx.Get("UrlParams").(string)

	ipAddress := ctx.RemoteIP().String()

	ctx.Request.Header.VisitAll(func(key, value []byte) {
		if string(key) == "X-Forwarded-For" {
			ipAddress = string(value)
		}
	})
	return nil
}
```