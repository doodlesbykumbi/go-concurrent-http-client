# concurrent-http-client

An exploration of concurrency within the context of faulty networks, and resource and rate limits. Retry and concurrency limits are offered as potential solutions.

TODO: provide examples of the faulty network issue

## Exploration

This section requires Docker.

In this section you'll explore the impact of "retry" vs "concurrency limits". To do that we'll run an http server fronted by an haproxy instance that has rate limiting enabled. The haproxy instance will allow a certain number of requests to go through per second, the goal is to explore the different approaches to having concurrent clients that are resilient to this rate limit.

The setup looks like this:
```
[server] <--> [haproxy with rate limit] <--> [concurrent client requests]
```

Any commands you run will assume that you've sourced `not-Makefile.sh` and that the current working directory is the root of this repository. 
```console
$ source ./not-Makefile.sh
```

### Setup

In a terminal, run `setup`. This will start haproxy with a particular rate limit on the TCP listener, and set up a container where you can run the test-server and use the test-client.
```console
$ RATE_LIMIT=3 setup
All set.
```

### Run server

In the same terminal as the previous step, run the server with a request duration of 2 (that means each request will take 2 seconds to get a response):
```console
$ run_server -request-duration 2
2020/03/23 17:27:14 Running server at :8080
```

### Run client

In a new terminal we can now play around with the client.

#### No limits

Start by running a client with no concurrency limits, but matching the RATE_LIMIT set above:
```console
$ run_client -num-requests 3
2020/03/23 17:36:20 2: starting
2020/03/23 17:36:20 0: starting
2020/03/23 17:36:20 1: starting
2020/03/23 17:36:21 .
2020/03/23 17:36:22 .
2020/03/23 17:36:22 2: code: 200
2020/03/23 17:36:22 1: code: 200
2020/03/23 17:36:22 0: code: 200
```

Notice that all three requests start at the same time. 2 seconds elapse, then all get 200 responses back. This makes sense because we're within the rate limit.

Run again but this time increasing the number of requests:
```console
$ run_client -num-requests 3
2020/03/23 17:40:13 3: starting
2020/03/23 17:40:13 0: starting
2020/03/23 17:40:13 1: starting
2020/03/23 17:40:13 2: starting
2020/03/23 17:40:13 2: err: Get http://host.docker.internal:8081: EOF
2020/03/23 17:40:14 .
2020/03/23 17:40:15 .
2020/03/23 17:40:15 0: code: 200
2020/03/23 17:40:15 1: code: 200
2020/03/23 17:40:15 3: code: 200
```

As before, all requests start at the same time. 1 immediately fails because of the rate limit. 2 seconds elapse, then the ones that got through get 200 responses back.

#### Max concurrency limits

Now we limit the number of concurrent requests, and number of concurrent requests per second. Truth be told these should really be separated. There's a distinction between limiting the number of things you can do per second, and limiting the number of things you can be doing at the same time. For this example I made that number the same.

```console
$ run_client -num-requests 4 -max-concurrency 3
2020/03/23 17:49:38 0: starting
2020/03/23 17:49:38 1: starting
2020/03/23 17:49:38 2: starting
2020/03/23 17:49:39 .
2020/03/23 17:49:40 .
2020/03/23 17:49:40 0: code: 200
2020/03/23 17:49:40 3: starting
2020/03/23 17:49:40 1: code: 200
2020/03/23 17:49:40 2: code: 200
2020/03/23 17:49:41 .
2020/03/23 17:49:42 .
2020/03/23 17:49:42 3: code: 200
```

Notice that in the first instance 3 requests start at the same time. 2 seconds elapse, those first 3 get 200 responses back, then the last request starts. Another 2 seconds elapse, the last gets a 200 responses back. This is because we're rate limiting ourselves to ensure we're only ever doing 3 things per second.

#### Retry

Now let's go back to unlimited concurrency

```console
$ run_client -num-requests 4 -retry
2020/03/23 17:47:31 3: starting
2020/03/23 17:47:31 1: starting
2020/03/23 17:47:31 2: starting
2020/03/23 17:47:31 0: starting
2020/03/23 17:47:32 .
2020/03/23 17:47:33 retry 1
2020/03/23 17:47:33 .
2020/03/23 17:47:34 3: code: 200
2020/03/23 17:47:34 0: code: 200
2020/03/23 17:47:34 2: code: 200
2020/03/23 17:47:34 .
2020/03/23 17:47:35 .
2020/03/23 17:47:36 .
2020/03/23 17:47:37 .
2020/03/23 17:47:38 .
2020/03/23 17:47:39 retry 2
2020/03/23 17:47:39 .
2020/03/23 17:47:40 .
2020/03/23 17:47:41 1: code: 200
```

Notice that all requests start at the same time. 1 second later a retry occurs. An additional second elapses, those first 3 get 200 responses back, another retry occurs. Another 2 seconds elapse, the remaining request gets a 200 responses back. We start off by hammering the server, and some requests don't succeed. We proceed to retry them following the retry policy (which is currently hardcoded).

#### Some thoughts

Rate limiting, concurrency limiting and retry can be used in tandem... lol I'll update this later.


