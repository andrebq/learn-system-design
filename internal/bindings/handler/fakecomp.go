package handler

import (
	"context"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func FakeComputations(ctx context.Context) func(*lua.LState) int {
	return func(L *lua.LState) int {
		// register functions to the table
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"slow": func(L *lua.LState) int {
				seconds := time.Duration(L.CheckNumber(1)) * time.Second
				select {
				case <-ctx.Done():
					L.Push(lua.LFalse)
				case <-time.After(seconds):
					L.Push(lua.LTrue)
				}
				return 1
			},
		})

		// returns the module
		L.Push(mod)
		return 1
	}
}
