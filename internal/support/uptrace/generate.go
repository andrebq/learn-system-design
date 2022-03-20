//go:generate statik -src=. -include=*.yaml -ns=github.com/andrebq/learn-system-design/internal/support/uptrace

package uptrace

import (
	_ "github.com/andrebq/learn-system-design/internal/support/uptrace/statik"
)
