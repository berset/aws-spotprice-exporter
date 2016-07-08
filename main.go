package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	addr        = flag.String("listen-address", ":9090", "The address to listen on for HTTP requests.")
	sleep       = flag.Int("sleep", 15, "sleep between polls")
	aws_az      = flag.String("az", "eu-west-1", "AWS AvailabilityZone")
	price_gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "spot_price",
		Help: "Spot price of aws instance type.",
	}, []string{"AvailabilityZone", "InstanceType", "ProductDescription"})
)

func init() {
	prometheus.MustRegister(price_gauge)
}

func main() {
	flag.Parse()

	go func() {
		for {
			fetch_prices()
			time.Sleep(time.Duration(*sleep) * time.Second)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", prometheus.Handler())
	http.ListenAndServe(*addr, nil)
}

func fetch_prices() {
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String(*aws_az)})
	params := &ec2.DescribeSpotPriceHistoryInput{}

	resp, err := svc.DescribeSpotPriceHistory(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	prices := resp.SpotPriceHistory
	for _, price := range prices {
		spot_p, err := strconv.ParseFloat(*price.SpotPrice, 64)
		if err == nil {
			price_gauge.WithLabelValues(
				*price.AvailabilityZone,
				*price.InstanceType,
				*price.ProductDescription).Set(spot_p)
		} else {
			fmt.Println(err)
		}
	}
}
