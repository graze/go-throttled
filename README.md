# HTTP Throttled Transport

An HTTP transport that throttles requests using a RateLimiter. RoundTripper middleware so that you can use it with
third party clients

## Usage

```go
client := &http.Client{
    Transport: throlled.NewTransport(http.DefaultTransport, rate.NewLimiter(rate.Limit(10), 1)),
}

// query is now rate limited to 10/second
res, err := client.Get("/bla")
```

Using a rate limiter with a bucket

```go
client := &http.Client{
    Transport: throlled.NewTransport(http.DefaultTransport, rate.NewLimiter(rate.Limit(4), 40)),
}

// query is now rate limited to 4/second with a bucket of 40
res, err := client.Get("/bla")
```
