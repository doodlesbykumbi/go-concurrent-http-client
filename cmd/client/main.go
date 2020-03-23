package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

func main() {
	var maxConcurrency int
	var numTasks int
	var retry bool
	var url string
	flag.IntVar(&maxConcurrency, "max-concurrency", math.MaxInt64, "max concurrency")
	flag.IntVar(&numTasks, "num-requests", 10, "num requests")
	flag.BoolVar(&retry, "retry", false, "retry")
	flag.StringVar(&url, "url", "http://host.docker.internal:8081", "url")
	flag.Parse()

	client := &http.Client{
		Transport: http.DefaultTransport,
	}

	if retry {
		retriableClient := &retryablehttp.Client{
			HTTPClient:   client,
			RetryWaitMax: 3 * time.Second,
			Backoff:      retryablehttp.LinearJitterBackoff,
			CheckRetry:   retryablehttp.DefaultRetryPolicy,
			RetryMax:     3,
			RequestLogHook: func(logger retryablehttp.Logger, request *http.Request, i int) {
				if i > 0 {
					log.Println("retry " + fmt.Sprintf("%d", i))
				}
			},
		}

		client = WrappedRetriableHTTPClient(retriableClient)
	}

	// per second ticker
	go func() {
		ticker := time.Tick(1 * time.Second)
		for {
			<-ticker
			log.Printf(".")
		}
	}()

	var concurrencyLimitOn = maxConcurrency != math.MaxInt64

	// wg makes sure you know when all the running tasks are done
	var wg sync.WaitGroup
	// ensures you start maxConcurrency tasks per second
	var limiter *rate.Limiter
	// sem makes sure you have up to maxConcurrency tasks running at any given time
	var sem chan bool

	if concurrencyLimitOn {
		limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(maxConcurrency)), 1)
		sem = make(chan bool, maxConcurrency)
	}

	add  := func() {
		if concurrencyLimitOn {
			err := limiter.Wait(context.Background())
			if err != nil {
				panic(err)
			}
			sem <- true
		}

		wg.Add(1)
	}

	done := func () {
		if concurrencyLimitOn {
			<-sem
		}

		wg.Done()
	}
	wait := func () {
		wg.Wait()
	}


	// task runner
	for j := 0; j < numTasks; j++ {
		i := j

		add()

		go func() {
			defer done()
			log.Printf("%d: starting", i)

			request, _ := http.NewRequest(http.MethodGet, url, nil)

			res, err := client.Do(request)
			out := ""
			if err != nil {
				out += "err: " + err.Error()
			} else {
				out += "code: " + fmt.Sprintf("%d", res.StatusCode)
			}

			log.Printf("%d: %s", i, out)
		}()
	}

	wait()
}
