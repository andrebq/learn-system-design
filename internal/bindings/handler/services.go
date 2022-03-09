package handler

import (
	"bytes"
	"context"
	"io/ioutil"
	"math/rand"
	"net/http"

	"github.com/andrebq/learn-system-design/control"
	lua "github.com/yuin/gopher-lua"
)

func ServicesLoader(ctx context.Context, options []*control.Server) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		// register functions to the table
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"call": func(L *lua.LState) int {
				var name, body string
				name = L.CheckString(1)
				if L.GetTop() > 1 {
					body = L.CheckString(2)
				}
				server := randomOptionByName(options, name)
				if server == nil {
					L.RaiseError("handler: unable to find any server that implements service %v", name)
					return 0
				}
				req, err := http.NewRequestWithContext(ctx, "POST", server.Endpoint, bytes.NewBufferString(body))
				if err != nil {
					L.RaiseError("handler: unable create POST request on %v for service %v", server.Endpoint, body)
					return 0
				}
				// TODO: add open-telemetry headers later on
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					L.RaiseError("handler: unable perform POST request on %v for service %v", server.Endpoint, body)
					return 0
				}
				defer res.Body.Close()
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					L.RaiseError("handler: unable read response for POST on %v for service %v", server.Endpoint, body)
					return 0
				}
				L.Push(lua.LString(string(resBody)))
				return 1
			},
		})

		// returns the module
		L.Push(mod)
		return 1
	}
}

func randomOptionByName(options []*control.Server, name string) *control.Server {
	var validOptions []int
	for i, v := range options {
		if v.Name == name {
			validOptions = append(validOptions, i)
		}
	}
	switch len(validOptions) {
	case 0:
		return nil
	case 1:
		return options[validOptions[0]]
	}
	idx := rand.Intn(len(validOptions))
	return options[validOptions[idx]]
}
