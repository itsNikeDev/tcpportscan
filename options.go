package tcpportscan

import (
	"fmt"
	"time"
)

// Scanner проверяет TCP-порты на одном или нескольких хостах.
//
// Создаётся через New. Scanner не содержит изменяемого состояния
// и безопасен для конкурентного использования: один экземпляр можно
// переиспользовать для любого количества сканирований, в том числе
// параллельных.
type Scanner struct {
	concurrency int
	timeout     time.Duration
	resolver    Resolver
}

// Option настраивает Scanner. Опции применяются в New.
type Option func(*Scanner) error

// New создаёт Scanner с заданными опциями.
//
// По умолчанию выполняется до 100 одновременных попыток подключения
// с таймаутом 1 секунда, DNS-имена резолвятся через net.DefaultResolver.
func New(options ...Option) (*Scanner, error) {
	s := &Scanner{
		concurrency: 100,
		timeout:     time.Second,
		resolver:    newNetResolver(),
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// WithConcurrency задаёт максимальное число одновременных попыток
// подключения. Значение должно быть положительным.
func WithConcurrency(n int) Option {
	return func(s *Scanner) error {
		if n <= 0 {
			return fmt.Errorf("concurrency must be positive")
		}

		s.concurrency = n
		return nil
	}
}

// WithConnectTimeout задаёт таймаут одной попытки TCP-подключения.
// Значение должно быть положительным.
func WithConnectTimeout(timeout time.Duration) Option {
	return func(s *Scanner) error {
		if timeout <= 0 {
			return fmt.Errorf("connect timeout must be positive")
		}

		s.timeout = timeout
		return nil
	}
}

// WithResolver задаёт резолвер для поиска IP-адресов по DNS-именам.
// Он не должен быть nil. По умолчанию используется net.DefaultResolver.
func WithResolver(resolver Resolver) Option {
	return func(s *Scanner) error {
		if resolver == nil {
			return fmt.Errorf("resolver must not be nil")
		}

		s.resolver = resolver
		return nil
	}
}
