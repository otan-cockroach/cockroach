package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:57800", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	t := time.Now()
	client := tspb.NewTimeSeriesClient(conn)
	resp, err := client.Query(context.Background(), &tspb.TimeSeriesQueryRequest{
		StartNanos: t.Add(-time.Minute).UnixNano(),
		EndNanos:   t.UnixNano(),
		Queries: []tspb.Query{
			{Name: "cr.node.sql.select.count"},
		},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("resp: %v\n", resp)
}
