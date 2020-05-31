package main

import (
	"fmt"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	sharedapp "local.packages/sharedapp"

	"github.com/gorilla/mux"
	_ "google.golang.org/grpc"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/integrations/nrlogrus"
	"github.com/sirupsen/logrus"
	_ "github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
)

var (
	coupon_url string
	log *logrus.Logger
)

func order(rw http.ResponseWriter, req *http.Request) {
	txn := newrelic.FromContext(req.Context())
	e := log.WithContext(req.Context())
	conn, err := grpc.Dial(
		coupon_url,
		grpc.WithInsecure(),
		// Add the New Relic gRPC client instrumentation
		grpc.WithUnaryInterceptor(nrgrpc.UnaryClientInterceptor),
		grpc.WithStreamInterceptor(nrgrpc.StreamClientInterceptor),
	)
	if err != nil {
		txn.NoticeError(err)
		e.Error(err)
		rw.Write([]byte("error"))
		return
	}
	defer conn.Close()

	couponId := req.URL.Query().Get("id")
	client := sharedapp.NewCouponApplicationClient(conn)
	coupon := &sharedapp.Coupon{
		Id: couponId,
	}
	res, err := client.Validate(req.Context(), coupon)
	if err != nil {
		txn.NoticeError(err)
		log.Print(err)
		rw.Write([]byte("error"))
		return
	}
	e.Infof("coupon利用 ID= %s", couponId)
	if res.IsValid {
		rw.Write([]byte(fmt.Sprintf("値引額: %v円", res.Amount)))
	} else {
		rw.Write([]byte("指定したクーポンは利用できません"))
	}
}

func async(rw http.ResponseWriter, req *http.Request) {
	txn := newrelic.FromContext(req.Context())
	e := log.WithContext(req.Context())
	wg := &sync.WaitGroup{}
	numTaks := 5
	for i := 0; i < numTaks; i++ {
		wg.Add(1)
		go func(idx int, txn *newrelic.Transaction) {
			defer wg.Done()
			defer txn.StartSegment("segment-"+strconv.Itoa(idx)).End()
			e.Infof("Task実行 ID= %v", idx)
			time.Sleep(20 * time.Millisecond)
		}(i, txn.NewGoroutine())
	}

	wg.Wait()
	rw.Write([]byte("async work complete"))
}

func makeHandler(text string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(text))
	})
}

func main() {
	log = logrus.New()
	// To enable New Relic log decoration, use the
	log.SetFormatter(nrlogrusplugin.ContextFormatter{})
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(os.Stdout)
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("lab2 webportal"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
		newrelic.ConfigDistributedTracerEnabled(true),
		nrlogrus.ConfigStandardLogger(),
	)
	if err != nil {
		log.Panic(err)
	}

	coupon_url = os.Getenv("COUPON_SVC_URL")
	if coupon_url == "" {
		coupon_url = "couponservice:8001"
	}


	err = app.WaitForConnection(10 * time.Second)
	if err != nil {
		log.Panic(err)
	}
	defer app.Shutdown(10 * time.Second)

	r := mux.NewRouter()
	r.Use(nrgorilla.Middleware(app))

	r.HandleFunc("/async", async)
	r.HandleFunc("/order", order)

	http.ListenAndServe(":8000", r)
}
