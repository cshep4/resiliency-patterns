package retry

//go:generate go run go.uber.org/mock/mockgen -source=internal/retry/retry.go -destination=internal/mocks/retry_mock.go -package=mocks OrderProcessor
