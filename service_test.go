package jrpc2server

import (
	"testing"

	"github.com/erikdubbelboer/fasthttp"
)

// DemoAPI area
type DemoAPI struct{}

// Test Method to test
func (h *DemoAPI) Test(ctx *fasthttp.RequestCtx, args *struct{ ID string }, reply *struct{ LogID string }) error {
	reply.LogID = args.ID
	return nil
}

func Test(t *testing.T) {

	api := NewServer()
	err := api.RegisterService(new(DemoAPI), "demo")

	if err != nil {
		t.Error(err)
	}

	reqHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/api":
			api.APIHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	go fasthttp.ListenAndServe(":8081", reqHandler)

}
