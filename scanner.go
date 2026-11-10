package tcpportscan

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

// Scan проверяет заданные порты на заданных хостах и передаёт результаты
// в возвращаемый канал.
//
// В качестве хостов принимаются IPv4-адреса, IPv6-адреса и DNS-имена.
// DNS-имя, которому соответствует несколько адресов, даёт результат для
// каждого из них. Повторяющиеся хосты (без учёта регистра) и повторяющиеся
// адреса проверяются один раз. Порт 0 отклоняется; для построения
// проверенных списков портов из отдельных значений и диапазонов
// используйте Ports.
//
// Проверки выполняются конкурентно, с ограничением WithConcurrency.
// Отмена ctx останавливает запуск новых проверок и даёт уже запущенным
// завершиться; канал закрывается после доставки всех результатов.
// Вызывающая сторона должна читать канал до закрытия.
//
// Scan возвращает ошибку, ничего не запуская, если аргументы некорректны
// или ctx уже завершён.
func (s *Scanner) Scan(
	ctx context.Context,
	hosts []string,
	ports []uint16,
) (<-chan Result, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(hosts) == 0 {
		return nil, fmt.Errorf("hosts are empty")
	}

	if len(ports) == 0 {
		return nil, fmt.Errorf("ports are empty")
	}

	for _, port := range ports {
		if port == 0 {
			return nil, fmt.Errorf("invalid port: 0")
		}
	}

	jobs, err := buildJobs(ctx, s.resolver, hosts, ports)
	if err != nil {
		return nil, err
	}

	jobsCh := make(chan job)
	resultsCh := make(chan Result)

	wg := startWorkers(
		ctx,
		s.concurrency,
		jobsCh,
		resultsCh,
		s.timeout,
		checkPort,
	)

	go produceJobs(ctx, jobs, jobsCh)

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	return resultsCh, nil
}

func checkPort(
	ctx context.Context,
	host string,
	ip net.IP,
	port uint16,
	timeout time.Duration,
) (result Result) {
	result = Result{
		Host: host,
		IP:   ip,
		Port: port,
	}

	start := time.Now()

	defer func() {
		result.Duration = time.Since(start)
	}()

	address := net.JoinHostPort(ip.String(), strconv.Itoa(int(port)))

	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		result.Err = err
		result.State = classifyError(ctx, err)
		return result
	}

	conn.Close()

	result.State = StateOpen
	return
}
