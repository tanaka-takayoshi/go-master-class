package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	_ "database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	newrelic "github.com/newrelic/go-agent/v3/newrelic"
	"github.com/newrelic/go-agent/v3/integrations/logcontext/nrlogrusplugin"
	"github.com/newrelic/go-agent/v3/integrations/nrlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	sharedapp "local.packages/sharedapp"
	_ "github.com/newrelic/go-agent/v3/integrations/nrmysql"
)

var (
	dbx *sqlx.DB
	app *newrelic.Application
	log *logrus.Logger
)

type CouponDiscount struct {
	CouponType    string    `json:"coupon_type" db:"coupon_type"`
	Amount int64   `json:"amount" db:"amount"`
}

// Server is a gRPC server.
type CouponApplicationServer struct{}

func (c CouponApplicationServer) Validate(ctx context.Context, coupon *sharedapp.Coupon) (*sharedapp.CouponResult, error) {
	e := log.WithContext(ctx)
	defer newrelic.FromContext(ctx).StartSegment("validate").End()

	var couponDiscount CouponDiscount
	query := "SELECT * FROM coupon_master WHERE coupon_type=?"

	e.Debugf("coupon情報を取得 ID=%s", coupon.Id)
	// From
	err := dbx.GetContext(ctx, &couponDiscount, query, coupon.Id)
	if (err != nil) {
		if err == sql.ErrNoRows {
			return &sharedapp.CouponResult{IsValid: false, Amount: 0, Message: "coupon ID is not valid."}, nil
		}
		return nil, err
	}

	return &sharedapp.CouponResult{IsValid: true, Amount: couponDiscount.Amount, Message: ""}, nil
}

func main() {
	log = logrus.New()
	// To enable New Relic log decoration, use the
	log.SetFormatter(nrlogrusplugin.ContextFormatter{})
	log.SetLevel(logrus.TraceLevel)
	log.SetOutput(os.Stdout)
	var err error

	host := os.Getenv("MYSQL_HOSTNAME")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = "3306"
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		port = "3306"
	}
	user := os.Getenv("MYSQL_USER")
	if user == "" {
		user = "lab2"
	}
	dbname := os.Getenv("MYSQL_DATABASE")
	if dbname == "" {
		dbname = "lab2"
	}
	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = "password"
	}

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		user,
		password,
		host,
		port,
		dbname,
	)
	log.Info(dsn)

	dbx, err = sqlx.Open("nrmysql", dsn)
	if err != nil {
		log.Panicf("failed to connect to DB: %s.", err.Error())
	}
	defer dbx.Close()

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName("lab2 couponservice"),
		newrelic.ConfigLicense(os.Getenv("NEW_RELIC_LICENSE_KEY")),
		newrelic.ConfigDistributedTracerEnabled(true),
		nrlogrus.ConfigStandardLogger(),
	)
	if nil != err {
		log.Panic(err)
	}

	lis, err := net.Listen("tcp", "0.0.0.0:8001")
	if err != nil {
		log.Panic(err)
	}
	grpcServer := grpc.NewServer(
		// Add the New Relic gRPC server instrumentation
		grpc.UnaryInterceptor(nrgrpc.UnaryServerInterceptor(app)),
		grpc.StreamInterceptor(nrgrpc.StreamServerInterceptor(app)),
	)
	sharedapp.RegisterCouponApplicationServer(grpcServer, &CouponApplicationServer{})
	grpcServer.Serve(lis)
}

