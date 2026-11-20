package ui

type ReadingStatus uint8

const (
	ReadingStatusWaiting ReadingStatus = iota
	ReadingStatusUnavailable
	ReadingStatusCurrent
	ReadingStatusStale
)

type Reading[T any] struct {
	Value  T
	Status ReadingStatus
}

func (r Reading[T]) HasValue() bool {
	return r.Status == ReadingStatusCurrent || r.Status == ReadingStatusStale
}

func ReadingText[T any](reading Reading[T], format func(T) string) string {
	switch reading.Status {
	case ReadingStatusWaiting:
		return "Waiting for sample"
	case ReadingStatusUnavailable:
		return "Unavailable"
	case ReadingStatusStale:
		return format(reading.Value) + " (stale)"
	case ReadingStatusCurrent:
		return format(reading.Value)
	default:
		return "Unavailable"
	}
}
