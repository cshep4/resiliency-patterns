package circuitbreaker

//go:generate go run go.uber.org/mock/mockgen -source=internal/circuitbreaker/circuitbreaker.go -destination=internal/mocks/circuitbreaker_mock.go -package=mocks PaymentProcessor
