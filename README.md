# HTTP Throttled Transport

An HTTP transport that throttles requests using a RateLimiter. RoundTripper middleware so that you can use it with
third party clients

## Usage

```go
client := &http.Client{
    Transport: throttled.NewTransport(http.DefaultTransport, rate.NewLimiter(rate.Limit(10), 1)),
}

// requests are now rate limited to 10/second
res, err := client.Get("/bla")
```

Using a rate limiter with a bucket

```go
client := throttled.Client(rate.NewLimiter(rate.Limit(4), 40))

// requests are now rate limited to 4/second with a bucket of 40
res, err := client.Get("/bla")
```

You can wrap an existing client

```go
client = throttled.WrapClient(client, rate.NewLimiter(rate.Limit(4), 40))

// requests of an existing client are now rate limited to 4/second with a bucket of 40
res, err := client.Get("/bla")
```

## Development

You can test using the `make` method which will call go locally or use docker if not installed

```shell
make test
```
