package main

import (
	"fmt"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"sync"
	"time"

	sharedapp "local.packages/sharedapp"

	"github.com/gorilla/mux"
	_ "google.golang.org/grpc"
	"github.com/sirupsen/logrus"
)

var (
	coupon_url string
	log *logrus.Logger
)

func order(rw http.ResponseWriter, req *http.Request) {
	conn, err := grpc.Dial(
		coupon_url,
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
		rw.Write([]byte("error"))
		return
	}
	log.Infof("coupon利用 ID= %s", couponId)
	if res.IsValid {
		rw.Write([]byte(fmt.Sprintf("値引額: %v円", res.Amount)))
	} else {
		rw.Write([]byte("指定したクーポンは利用できません"))
	}
}

func async(rw http.ResponseWriter, req *http.Request) {
	wg := &sync.WaitGroup{}
	numTaks := 5
	for i := 0; i < numTaks; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			time.Sleep(20 * time.Millisecond)
		}(i)
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
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(os.Stdout)

	coupon_url = os.Getenv("COUPON_SVC_URL")
	if coupon_url == "" {
		coupon_url = "couponservice:8001"
	}

	r := mux.NewRouter()

	r.HandleFunc("/async", async)
	r.HandleFunc("/order", order)

	http.ListenAndServe(":8000", r)
}
