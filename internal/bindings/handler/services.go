package handler

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/rs/zerolog/log"
	lua "github.com/yuin/gopher-lua"
)

func ServicesLoader(ctx context.Context) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		// register functions to the table
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"call": func(L *lua.LState) int {
				var name, body string
				name = L.CheckString(1)
				if L.GetTop() > 1 {
					body = L.CheckString(2)
				}

				endpoint := fmt.Sprintf("http://%v/", name)
				req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBufferString(body))
				if err != nil {
					log.Error().Err(err).Send()
					L.RaiseError("handler: create request on %v for service %v", endpoint, name)
					return 0
				}
				res, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Error().Err(err).Send()
					L.RaiseError("handler: unable perform POST request on %v for service %v", endpoint, name)
					return 0
				}
				defer res.Body.Close()
				resBody, err := ioutil.ReadAll(res.Body)
				if err != nil {
					L.RaiseError("handler: unable read response for POST on %v for service %v", endpoint, body)
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
