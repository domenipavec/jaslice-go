package luna

import (
	"time"
)

const (
	lunarCycle = 2551442 * time.Second
)

func (luna *Luna) getPhase() byte {
	d := time.Since(time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC))

	return byte(((((d%lunarCycle)+12*time.Hour)/(24*time.Hour) + 1) / 3) % 10)
}
