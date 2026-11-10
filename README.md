# tcpportscan

Go-пакет для проверки доступности TCP-портов (connect-сканирование) на одном или нескольких хостах. Только стандартная библиотека.

## Использование

```go
scanner, err := tcpportscan.New(
    tcpportscan.WithConcurrency(100),
    tcpportscan.WithConnectTimeout(500*time.Millisecond),
)
if err != nil {
    return err
}

ports, err := tcpportscan.Ports(22, 80, tcpportscan.Range(8000, 8080))
if err != nil {
    return err
}

results, err := scanner.Scan(ctx, []string{"192.168.1.10"}, ports)
if err != nil {
    return err
}

for result := range results {
    fmt.Printf("%s:%d %s\n", result.Host, result.Port, result.State)
}
```

Результаты приходят в канал по мере готовности, порядок не гарантируется. Канал закрывается, когда все проверки завершены или контекст отменён, — его нужно читать до конца.

## API

- `New(options ...Option) (*Scanner, error)` — создаёт сканер. Экземпляр можно переиспользовать и вызывать из разных горутин.
- Опции: `WithConcurrency` (максимум одновременных подключений, по умолчанию 100), `WithConnectTimeout` (таймаут одного подключения, по умолчанию 1 сек), `WithResolver` (свой DNS-резолвер, по умолчанию `net.DefaultResolver`).
- `Ports(values ...any) ([]uint16, error)` — собирает список портов из отдельных значений и диапазонов `Range(from, to)`. Дубликаты убираются; порт 0, порт больше 65535 и обратный диапазон — ошибка.
- `Scan(ctx, hosts, ports) (<-chan Result, error)` — запускает сканирование. Хосты: IPv4, IPv6 или DNS-имя; если имя резолвится в несколько IP, проверяются все. Дубликаты хостов убираются. Пустой хост, неразрешимое имя, пустой список портов — ошибка.
- `Result{Host, IP, Port, State, Duration, Err}` — результат по каждому порту. Состояния: `open`, `closed`, `timeout`, `unreachable`, `canceled`, `error`.

## Тесты

```
go test -race ./...
```
