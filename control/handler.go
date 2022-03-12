package control

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/andrebq/learn-system-design/internal/mutex"
	"github.com/andrebq/learn-system-design/internal/render"
	"github.com/julienschmidt/httprouter"
)

type (
	control struct {
		globalLock mutex.Zone
		stressors  *stressorList
		services   *serviceList
		instances  *instanceList
	}

	instanceList struct {
		items map[string]*Instance
	}

	stressorList struct {
		items []*Stressor
	}

	serviceList struct {
		items []*Server
	}

	Instance struct {
		Name                string            `json:"name"`
		LastPing            time.Time         `json:"lastPing,omitempty"`
		TimeSinceLastPingMs int64             `json:"timeSinceLastPingMs,omitempty"`
		Services            map[string]string `json:"services"`
		Metrics             struct {
			Requests int64 `json:"requests"`
		} `json:"metrics"`
	}

	Stressor struct {
		BaseEndpoint string `json:"baseEndpoint"`
		Name         string `json:"name"`
	}

	Server struct {
		Service  string `json:"service"`
		Endpoint string `json:"endpoint"`
	}
)

func Handler() http.Handler {
	r := httprouter.New()
	c := &control{
		services:  &serviceList{},
		stressors: &stressorList{},
		instances: &instanceList{
			items: make(map[string]*Instance),
		},
	}
	r.HandlerFunc("PUT", "/register/service/:service", c.registerServer)
	r.HandlerFunc("PUT", "/register/stressor/:name", c.registerStressor)
	r.HandlerFunc("PUT", "/register/instance/:name", c.registerInstance)
	r.HandlerFunc("GET", "/registry", c.getRegistry)
	r.HandlerFunc("GET", "/", c.getRegistry)
	return r
}

func (c *control) registerServer(rw http.ResponseWriter, req *http.Request) {
	params := httprouter.ParamsFromContext(req.Context())
	service := params.ByName("service")
	r := Server{}
	if err := render.ReadJSONOrFail(rw, req, &r); err != nil {
		return
	}
	r.Service = service
	r.Endpoint = strings.TrimRight(r.Endpoint, "/")
	if _, err := url.Parse(r.Endpoint); err != nil || len(r.Endpoint) == 0 {
		render.WriteError(rw, http.StatusBadRequest, "invalid server endpoint")
		return
	}
	mutex.Run(c.globalLock.Exclusive(), func() {
		c.services.addServer(r)
	})
	render.WriteSuccess(rw, http.StatusOK, "Server added to the list")
}

func (c *control) registerStressor(rw http.ResponseWriter, req *http.Request) {
	name := httprouter.ParamsFromContext(req.Context()).ByName("name")
	r := Stressor{}
	if err := render.ReadJSONOrFail(rw, req, &r); err != nil {
		return
	}
	r.Name = name
	r.BaseEndpoint = strings.TrimRight(r.BaseEndpoint, "/")
	if _, err := url.Parse(r.BaseEndpoint); err != nil || len(r.BaseEndpoint) == 0 {
		render.WriteError(rw, http.StatusBadRequest, "Invalid endpoint")
	}
	mutex.Run(c.globalLock.Exclusive(), func() {
		c.stressors.addStressor(r)
	})
	render.WriteSuccess(rw, http.StatusOK, "Stressor added to the list")
}

func (c *control) registerInstance(rw http.ResponseWriter, req *http.Request) {
	name := httprouter.ParamsFromContext(req.Context()).ByName("name")
	i := Instance{}
	if err := render.ReadJSONOrFail(rw, req, &i); err != nil {
		return
	}
	i.Name = name
	i.LastPing = time.Now()
	i.TimeSinceLastPingMs = 0
	mutex.Run(c.globalLock.Exclusive(), func() {
		c.instances.addInstance(i)
	})
	render.WriteSuccess(rw, http.StatusOK, "Instance added to the list")
}

func (c *control) getRegistry(rw http.ResponseWriter, req *http.Request) {
	var buf []byte
	var err error
	mutex.Run(c.globalLock.Exclusive(), func() {
		c.instances.trim()
	})
	mutex.Run(c.globalLock.Shared(), func() {
		buf, err = json.Marshal(struct {
			Servers   []*Server            `json:"servers"`
			Stressors []*Stressor          `json:"stressor"`
			Instances map[string]*Instance `json:"instances"`
		}{
			Servers:   c.services.items,
			Stressors: c.stressors.items,
			Instances: c.instances.items,
		})
	})
	if err != nil {
		render.WriteError(rw, http.StatusInternalServerError, "Bad server, could not handle the request")
		return
	}
	render.WriteJSONRaw(rw, http.StatusOK, buf)
}

func (sl *stressorList) addStressor(s Stressor) {
	for _, v := range sl.items {
		if v.BaseEndpoint == s.BaseEndpoint {
			return
		}
		v.Name = s.Name
	}
	sl.items = append(sl.items, &s)
}

func (sl *serviceList) addServer(s Server) {
	for _, v := range sl.items {
		if v.Service == s.Service && v.Endpoint == s.Endpoint {
			return
		}
	}
	sl.items = append(sl.items, &s)
}

func (il *instanceList) addInstance(i Instance) {
	il.items[i.Name] = &i
	il.trim()
}

func (il *instanceList) trim() {
	now := time.Now()
	for idx, v := range il.items {
		v.TimeSinceLastPingMs = now.Sub(v.LastPing).Milliseconds()
		if v.TimeSinceLastPingMs < 0 {
			v.TimeSinceLastPingMs = 0
			continue
		} else if v.TimeSinceLastPingMs > time.Minute.Milliseconds() {
			delete(il.items, idx)
		}
	}

}
