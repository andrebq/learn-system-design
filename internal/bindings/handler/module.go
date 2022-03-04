package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/andrebq/learn-system-design/internal/logutil"
	lua "github.com/yuin/gopher-lua"
)

func Loader(req *http.Request, res http.ResponseWriter) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		// register functions to the table
		mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
			"req": func(L *lua.LState) int {
				ud := L.NewUserData()
				ud.Value = req
				L.Push(ud)
				return 1
			},
			"addHeader": func(L *lua.LState) int {
				key := L.CheckString(1)
				value := L.CheckString(2)
				res.Header().Add(key, value)
				return 0
			},
			"writeStatus": func(L *lua.LState) int {
				status := L.CheckInt(1)
				res.WriteHeader(status)
				return 0
			},
			"writeBody": func(L *lua.LState) int {
				body := L.CheckString(1)
				io.WriteString(res, body)
				return 0
			},
			"log": func(L *lua.LState) int {
				log := logutil.Acquire(req.Context())
				ev := log.Info()
				for i := 0; i <= L.GetTop(); i++ {
					val := L.CheckAny(i)
					switch val.Type() {
					case lua.LTUserData:
						ev = ev.Str(strconv.Itoa(i), fmt.Sprintf("%v", val.(*lua.LUserData).Value))
					case lua.LTNil:
						continue
					default:
						ev = ev.Stringer(strconv.Itoa(i), val)
					}
				}
				ev.Send()
				return 0
			},
		})

		// returns the module
		L.Push(mod)
		return 1
	}
}
