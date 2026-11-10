package tcpportscan

import (
	"net"
	"time"
)

// State описывает итог проверки одного порта.
type State string

const (
	// StateOpen означает, что TCP-соединение установлено.
	StateOpen State = "open"

	// StateClosed означает, что соединение явно отклонено.
	StateClosed State = "closed"

	// StateTimeout означает, что подключение не завершилось за
	// заданный таймаут.
	StateTimeout State = "timeout"

	// StateUnreachable означает, что узел или сеть недоступны.
	StateUnreachable State = "unreachable"

	// StateCanceled означает, что проверка отменена через контекст.
	StateCanceled State = "canceled"

	// StateError означает, что проверка завершилась любой другой
	// ошибкой.
	StateError State = "error"
)

// Result — результат проверки одного порта на одном хосте.
type Result struct {
	// Host — цель в том виде, в каком её передали в Scan.
	Host string

	// IP — разрешённый адрес, к которому выполнялось подключение.
	IP net.IP

	// Port — проверенный TCP-порт.
	Port uint16

	// State — итог проверки.
	State State

	// Duration — длительность попытки подключения.
	Duration time.Duration

	// Err — исходная ошибка. Равна nil, только если State равен
	// StateOpen.
	Err error
}
