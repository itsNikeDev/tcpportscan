// Package tcpportscan предоставляет конкурентное TCP Connect-сканирование
// портов на одном или нескольких хостах.
//
// Scanner, созданный через New, резолвит каждый целевой хост, подключается
// ко всем указанным портам с ограниченным числом одновременных соединений
// и отправляет структурированный Result по каждому порту в канал.
//
//	scanner, err := tcpportscan.New(
//		tcpportscan.WithConcurrency(100),
//		tcpportscan.WithConnectTimeout(500*time.Millisecond),
//	)
//	if err != nil {
//		return err
//	}
//
//	ports, err := tcpportscan.Ports(22, 80, tcpportscan.Range(8000, 8080))
//	if err != nil {
//		return err
//	}
//
//	results, err := scanner.Scan(ctx, []string{"192.168.1.10"}, ports)
//	if err != nil {
//		return err
//	}
//
//	for result := range results {
//		fmt.Println(result.Host, result.Port, result.State)
//	}
package tcpportscan
