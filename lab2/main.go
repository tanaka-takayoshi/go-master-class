package main

import (
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func async(rw http.ResponseWriter, req *http.Request) {
	wg := &sync.WaitGroup{}
	numTaks := 5
	for i := 0; i < numTaks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond)
		}()
	}

	wg.Wait()
	rw.Write([]byte("async work complete"))
}

func main() {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("lab2"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
		newrelic.ConfigDebugLogger(os.Stdout),
	)

	if err != nil {
		panic(err)
	}

	http.HandleFunc(newrelic.WrapHandleFunc(app, "/async", async))
	http.ListenAndServe(":8080", nil)

}
