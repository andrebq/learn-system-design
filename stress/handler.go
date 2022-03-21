package stress

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/andrebq/learn-system-design/control"
	"github.com/andrebq/learn-system-design/internal/logutil"
	"github.com/andrebq/learn-system-design/internal/render"
	"github.com/andrebq/learn-system-design/internal/router"
	vegeta "github.com/tsenart/vegeta/lib"
)

type (
	h struct {
		sync.Mutex

		testLock sync.Mutex

		ongoing bool

		test *StressTest

		hdrHistogram []byte
		status       []byte

		controlEndpoint string
		name            string
		publicEndpoint  string
	}

	StressTest struct {
		Name              string        `json:"name"`
		Target            string        `json:"target"`
		Method            string        `json:"method"`
		RequestsPerSecond int           `json:"requestsPerSecond"`
		Workers           int           `json:"workers"`
		Timeout           time.Duration `json:"timeout"`
		Sustain           time.Duration `json:"sustain"`
	}
)

func Handler(ctx context.Context, name string, controlEndpoint, publicEndpoint string) http.Handler {
	router := router.New()
	handler := &h{
		name:            name,
		publicEndpoint:  publicEndpoint,
		controlEndpoint: controlEndpoint,
	}
	router.HandlerFunc("GET", "/reports/hdr-histogram.txt", handler.getHDRHistogram)
	router.HandlerFunc("POST", "/start-test", handler.startTest)
	router.HandlerFunc("GET", "/", handler.getStatus)
	go handler.registration(ctx)
	_ = handler
	return router
}

func (h *h) getHDRHistogram(rw http.ResponseWriter, req *http.Request) {
	h.Lock()
	var buf []byte
	buf = append(buf, h.hdrHistogram...)
	h.Unlock()

	if len(buf) == 0 {
		http.Error(rw, "Data is not available yet, try again in a couple of seconds.", http.StatusConflict)
		return
	}
	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.Header().Add("Content-Length", strconv.Itoa(len(buf)))
	rw.WriteHeader(http.StatusOK)
	rw.Write(buf)
}

func (h *h) startTest(rw http.ResponseWriter, req *http.Request) {
	var test StressTest
	if err := render.ReadJSONOrFail(rw, req, &test); err != nil {
		return
	}
	if test.Sustain == 0 || test.Sustain > time.Minute {
		test.Sustain = time.Second * 5
	}
	if test.Timeout > test.Sustain || test.Timeout <= 0 {
		test.Timeout = test.Sustain
	}
	if test.Method == "" {
		test.Method = "GET"
	}
	if _, err := url.Parse(test.Target); err != nil || len(test.Target) == 0 {
		render.WriteError(rw, http.StatusBadRequest, "Invalid or missing target")
		return
	}
	if test.RequestsPerSecond <= 0 {
		test.RequestsPerSecond = 10
	}
	if test.Workers <= 0 {
		test.Workers = runtime.NumCPU()
	}

	h.Lock()
	defer h.Unlock()
	if h.ongoing {
		render.WriteError(rw, http.StatusConflict, "There is one test already in progress, try again later")
		return
	}

	h.test = &test
	h.ongoing = true
	go h.performTest(test)
	render.WriteSuccess(rw, http.StatusCreated, "Test in progress")
}

func (h *h) getStatus(rw http.ResponseWriter, req *http.Request) {
	var aux bytes.Buffer
	var status int
	h.Lock()
	switch {
	case !h.ongoing && len(h.status) == 0:
		io.WriteString(&aux, "no tests")
		io.WriteString(&aux, "\n")
		status = http.StatusOK
	case h.ongoing && len(h.status) == 0:
		io.WriteString(&aux, "... Test is in progress, partial results are not available")
		io.WriteString(&aux, "\n")
		status = http.StatusTooEarly
	case h.ongoing:
		io.WriteString(&aux, "... Test is in progress, partial results are partial")
		io.WriteString(&aux, "\n")
		aux.Write(h.status)
		status = http.StatusTooEarly
	default:
		aux.Write(h.status)
		status = http.StatusOK
	}
	h.Unlock()

	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.Header().Add("Content-Length", strconv.Itoa(aux.Len()))
	rw.WriteHeader(status)
	io.Copy(rw, &aux)
}

func (h *h) performTest(test StressTest) {
	defer func() {
		h.Lock()
		h.ongoing = false
		h.Unlock()
		h.notifyStatusChange(context.TODO())
	}()
	h.testLock.Lock()
	defer h.testLock.Unlock()

	a := vegeta.NewAttacker(vegeta.Workers(uint64(test.Workers)), vegeta.Timeout(test.Timeout))
	rate := vegeta.ConstantPacer{
		Freq: test.RequestsPerSecond,
		Per:  time.Second,
	}
	target := vegeta.NewStaticTargeter(vegeta.Target{
		Method: test.Method,
		URL:    test.Target,
	})
	results := a.Attack(target, rate, test.Sustain, test.Name)
	time.AfterFunc(test.Sustain, a.Stop)
	metrics := vegeta.Metrics{
		Histogram: &vegeta.Histogram{
			Buckets: vegeta.Buckets{
				time.Microsecond,
				time.Microsecond * 500,
				time.Millisecond,
				time.Millisecond * 10,
				time.Millisecond * 100,
				time.Millisecond * 500,
				time.Second,
				time.Second * 10,
				time.Second * 30,
				time.Minute,
			},
		},
	}
	i := 0
	for r := range results {
		i++
		metrics.Add(r)
		if i%100 == 0 {
			h.notifyStatusChange(context.TODO())
			h.reportResults(&metrics)
		}
	}
	h.reportResults(&metrics)

}

func (h *h) reportResults(m *vegeta.Metrics) {
	h.Lock()
	defer h.Unlock()

	r := vegeta.NewHDRHistogramPlotReporter(m)
	buf := bytes.Buffer{}
	r.Report(&buf)
	h.hdrHistogram = buf.Bytes()

	buf = bytes.Buffer{}
	r = vegeta.NewTextReporter(m)
	r.Report(&buf)
	h.status = buf.Bytes()
}

func (h *h) registration(ctx context.Context) {
	if h.controlEndpoint == "" {
		return
	}
	runtime.Gosched()
	sampled := logutil.Acquire(ctx) //.Sample(zerolog.Sometimes)
	tick := time.NewTicker(time.Second * 5)
	for {
		err := h.notifyStatusChange(ctx)
		if err != nil {
			sampled.Error().
				Str("control", h.controlEndpoint).
				Str("name", h.name).
				Str("endpoint", h.publicEndpoint).
				Err(err).
				Msg("Unable to register")
		}

		select {
		case <-tick.C:
		case <-ctx.Done():
			return
		}
	}
}

func (h *h) notifyStatusChange(ctx context.Context) error {
	return control.RegisterStressor(ctx, h.controlEndpoint, h.name, h.publicEndpoint, h.ongoing)
}
