package core

// RoundFromSeconds rounds the given timestamp (in seconds) up to the nearest interval.
func RoundFromSeconds(timestamp int64, interval int64) int64 {
	remainder := timestamp % interval

	if remainder == 0 {
		return timestamp
	}
	return timestamp + (interval - remainder)
}
