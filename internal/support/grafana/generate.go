//go:generate statik -src=. -include=*.yaml -ns=github.com/andrebq/learn-system-design/internal/support/grafana

package grafana

import (
	_ "github.com/andrebq/learn-system-design/internal/support/grafana/statik"
)
