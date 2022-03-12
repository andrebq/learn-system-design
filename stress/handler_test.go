package stress

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/steinfletcher/apitest"
)

func TestStress(t *testing.T) {
	var count int32
	underTest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		atomic.AddInt32(&count, 1)
	}))
	defer underTest.Close()

	target := StressTest{
		Name:              "test",
		Target:            underTest.URL,
		Method:            "GET",
		Workers:           1,
		Sustain:           time.Millisecond * 100,
		RequestsPerSecond: 10,
	}
	handler := Handler(context.TODO(), "test", "", "")
	apitest.Handler(handler).Get("/").Expect(t).Status(http.StatusOK).Body("no tests\n").End()
	apitest.Handler(handler).Post("/start-test").Body(toJson(t, target)).Expect(t).Status(http.StatusCreated).End()
	apitest.Handler(handler).Get("/").Expect(t).Status(http.StatusTooEarly).End()
	// multiply by 1.5 to account for the time required to compute the reports
	// in practice, this takes way less time
	time.Sleep(time.Duration(float64(target.Sustain) * 1.5))
	apitest.Handler(handler).Get("/").Expect(t).Status(http.StatusOK).End()
	t.Logf("Total number of calls: %v", count)
	if count <= 0 {
		t.Fatal("Handler under test was not called")
	}
}

func toJson(t *testing.T, body interface{}) string {
	buf, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf)
}
