# Best Practices

Collection of tips and tricks for working with LLM's in Go.

## Retry

OpenAI has a somewhat flaky rate limit, and it's easy to hit it. You can use a [retryablehttp](https://pkg.go.dev/github.com/hashicorp/go-retryablehttp) client to retry requests automatically:

```go
retryableClient := retryablehttp.NewClient()
retryableClient.RetryMax = 5
oaiCfg := openai.DefaultConfig(mustGetEnv("OPENAI_API_KEY"))
oaiCfg.HTTPClient = retryableClient.StandardClient()
oai := openai.NewClientWithConfig(oaiCfg)
```
