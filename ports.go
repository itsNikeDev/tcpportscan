package tcpportscan

import "fmt"

// PortRange описывает включающий диапазон TCP-портов.
type PortRange struct {
	From int
	To   int
}

// Range возвращает PortRange, покрывающий порты от from до to
// включительно. Диапазон валидируется при передаче в Ports.
func Range(from, to int) PortRange {
	return PortRange{
		From: from,
		To:   to,
	}
}

// Ports объединяет отдельные порты и диапазоны в один список без
// дубликатов. Отдельные порты передаются как int, диапазоны — как
// PortRange:
//
//	ports, err := tcpportscan.Ports(22, 80, tcpportscan.Range(8000, 8080))
//
// Каждый порт должен быть в диапазоне 1..65535, для каждого диапазона
// должно выполняться From <= To; при некорректном вводе возвращается
// ошибка. Дубликаты удаляются, порядок первого появления сохраняется.
func Ports(values ...any) ([]uint16, error) {
	var ports []uint16
	seen := make(map[uint16]struct{})

	addPort := func(port uint16) {
		if _, exists := seen[port]; exists {
			return
		}

		seen[port] = struct{}{}
		ports = append(ports, port)
	}

	for _, value := range values {
		switch v := value.(type) {
		case int:
			if v < 1 || v > 65535 {
				return nil, fmt.Errorf("invalid port: %d", v)
			}

			addPort(uint16(v))
		case PortRange:
			if v.From < 1 ||
				v.From > 65535 ||
				v.To < 1 ||
				v.To > 65535 ||
				v.From > v.To {
				return nil, fmt.Errorf(
					"invalid port range: %d-%d",
					v.From,
					v.To,
				)
			}

			for port := v.From; port <= v.To; port++ {
				addPort(uint16(port))
			}
		default:
			return nil, fmt.Errorf("unsupported port type: %T", value)
		}
	}
	return ports, nil
}
