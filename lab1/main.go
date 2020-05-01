package main

import (
	"fmt"
	"github.com/newrelic/go-agent/v3/newrelic"
	"math/rand"
	"net/http"
	"os"
	"time"
)

var (
	app *newrelic.Application
	client = &http.Client{
		Transport: newrelic.NewRoundTripper(nil),
	}
)

func index(rw http.ResponseWriter, req *http.Request) {
	txn := newrelic.FromContext(req.Context())
	txn.AddAttribute("customerLevel", req.URL.Query().Get("customerLevel"))

	time.Sleep(15 * time.Millisecond)
	rw.Write([]byte("Hello World"))
}

func hoge(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("hoge"))

	txn := newrelic.FromContext(req.Context())

	func() {
		defer txn.StartSegment("segment1").End()
		time.Sleep(100 * time.Millisecond)
	}()

	s2 := txn.StartSegment("segment2")
	time.Sleep(150 * time.Millisecond)
	s2.End()
}

func external(rw http.ResponseWriter, r *http.Request) {
	txn := newrelic.FromContext(r.Context())
	rw = txn.SetWebResponse(rw)
	rw.Write([]byte("外部呼び出し"))

	rand.Seed(time.Now().UnixNano())
	url := "http://example.com"
	if (rand.Intn(2) == 0) {
		url = "http://example.invalid"
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		txn.NoticeError(err)
		rw.Write([]byte(fmt.Sprintf("リクエスト作成失敗: %v", err)))
		return
	}

	req = newrelic.RequestWithTransactionContext(req, txn)
	//Transactionにアクセスする必要がない場合
	//req = req.WithContext(r.Context())

	//RoundTripperを設定せずにマニュアルで計測する場合、 segment.End() までのコメント解除
	//segment := newrelic.StartExternalSegment(txn, req)
	//client := http.Client{}
	resp, err := client.Do(req)
	//segment.End()

	if err != nil {
		txn.NoticeError(err)
		rw.Write([]byte(fmt.Sprintf("リクエスト実行失敗: %v", err)))
		return
	}
	resp.Body.Close()
	rw.Write([]byte(fmt.Sprintf("呼び出しステータス: %v", resp.Status)))
}

func instrumentHandler(name string, fn func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		txn := app.StartTransaction(name)
		defer txn.End()

		req = newrelic.RequestWithTransactionContext(req, txn)

		txn.SetWebRequestHTTP(req)
		rw = txn.SetWebResponse(rw)
		fn(rw, req)
	}
}

type replacementMux struct {
	app *newrelic.Application
	*http.ServeMux
}

func (mux *replacementMux) HandleFunc(pattern string, fn func(http.ResponseWriter, *http.Request)) {
	mux.ServeMux.HandleFunc(newrelic.WrapHandleFunc(mux.app, pattern, fn))
}

func main() {
	var err error
	app, err = newrelic.NewApplication(
		newrelic.ConfigAppName("lab1"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
		newrelic.ConfigDebugLogger(os.Stdout),
	)

	if (err != nil) {
		panic(err)
	}

	//mux := http.NewServeMux()
	//mux.HandleFunc("/", instrumentHandler("index", index))
	//mux.HandleFunc("/", instrumentHandler("/hoge", hoge))
	//mux.HandleFunc("/", instrumentHandler("/external", external))
	//
	//mux.HandleFunc(newrelic.WrapHandleFunc(app, "/", index))
	//mux.HandleFunc(newrelic.WrapHandleFunc(app, "/hoge", hoge))
	//mux.HandleFunc(newrelic.WrapHandleFunc(app, "/external", external))

	mux := replacementMux{ServeMux: http.NewServeMux(), app: app}
	mux.HandleFunc("/", index)
	mux.HandleFunc("/hoge", hoge)
	mux.HandleFunc("/external", external)

	http.ListenAndServe(":8123", mux)
}
