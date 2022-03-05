package stress

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
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
	}

	errorResponse struct {
		Msg    string `json:"msg"`
		Status int    `json:"status"`
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

func Handler() http.Handler {
	router := httprouter.New()
	handler := &h{}
	router.HandlerFunc("GET", "/reports/hdr-histogram.txt", handler.getHDRHistogram)
	router.HandlerFunc("POST", "/start-test", handler.startTest)
	router.HandlerFunc("GET", "/", handler.getStatus)
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
	if err := readJSONOrFail(rw, req, &test); err != nil {
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
		writeError(rw, http.StatusBadRequest, "Invalid or missing target")
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
		writeError(rw, http.StatusConflict, "There is one test already in progress, try again later")
		return
	}

	h.test = &test
	h.ongoing = true
	go h.performTest(test)
	writeSuccess(rw, http.StatusCreated, "Test in progress")
}

func (h *h) getStatus(rw http.ResponseWriter, req *http.Request) {
	var aux bytes.Buffer
	h.Lock()
	if h.ongoing {
		io.WriteString(&aux, "... Test is in progress, results are partial ...")
		io.WriteString(&aux, "\n")
	}
	aux.Write(h.status)
	h.Unlock()

	rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
	rw.Header().Add("Content-Length", strconv.Itoa(aux.Len()))
	rw.WriteHeader(http.StatusOK)
	io.Copy(rw, &aux)
}

func (h *h) performTest(test StressTest) {
	defer func() {
		h.Lock()
		h.ongoing = false
		h.Unlock()
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

func readJSONOrFail(rw http.ResponseWriter, req *http.Request, out interface{}) error {
	dec := json.NewDecoder(req.Body)
	err := dec.Decode(out)
	if err != nil {
		writeError(rw, http.StatusBadRequest, err.Error())
		return err
	}
	return err
}

func writeJSON(rw http.ResponseWriter, status int, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		http.Error(rw, `{"error":{"msg": "internal server error", "status": 500}}`, http.StatusInternalServerError)
		return err
	}
	rw.Header().Add("Content-Length", strconv.Itoa(len(data)))
	rw.Header().Add("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	rw.Write(data)
	return nil
}

func writeSuccess(rw http.ResponseWriter, status int, msg string) error {
	return writeJSON(rw, status, struct {
		OK  bool   `json:"ok"`
		Msg string `json:"msg"`
	}{
		OK:  true,
		Msg: msg,
	})
}

func writeError(rw http.ResponseWriter, status int, msg string) error {
	return writeJSON(rw, status, struct {
		Error errorResponse `json:"error"`
	}{
		Error: errorResponse{Msg: msg, Status: status},
	})
}
