package summary

import (
	"math"
	"time"
)

type Counter struct {
	Total     int
	MaxGap    int
	MaxRewind int
	Duplicate int

	startTime      time.Time
	firstTimestamp int
	lastTimestamp  int

	// for computing the cache content of live stream in server
	realTimeStart          time.Time
	realTimeCount          int
	estimatedCacheFps      float64
	cacheTimestampDuration int
	cacheDuration          time.Duration

	DiffThreshold int
}

func (c *Counter) TimestampDuration() int {
	return c.lastTimestamp - c.firstTimestamp
}

func (c *Counter) Duration() time.Duration {
	return time.Since(c.startTime)
}

func (c *Counter) EstimatedCacheFps() float64 {
	return c.estimatedCacheFps
}

func (c *Counter) Rate() float64 {
	return float64(c.Total) / float64(c.TimestampDuration()) * float64(1000)
}

func (c *Counter) RealRate() float64 {
	return float64(c.Total) / float64(c.Duration().Seconds())
}

func (c *Counter) CacheTimestampDuration() int {
	return c.cacheTimestampDuration
}

func (c *Counter) CacheDuration() time.Duration {
	return c.cacheDuration
}

func (c *Counter) Count(timestamp int) {
	if c.startTime.IsZero() {
		now := time.Now()
		c.startTime = now
		c.realTimeStart = now
	}
	if c.Total == 0 {
		c.firstTimestamp = timestamp
		c.lastTimestamp = timestamp
		c.Total++
		return
	}

	// compute real time fps
	if c.cacheTimestampDuration == 0 {
		c.realTimeCount++
		if time.Since(c.realTimeStart) >= time.Millisecond*100 {
			c.estimatedCacheFps = float64(c.realTimeCount) / time.Since(c.realTimeStart).Seconds()
			if math.Abs(float64(c.estimatedCacheFps)-c.Rate())/c.Rate()*100 < float64(c.DiffThreshold) {
				// dry of server cache
				c.cacheTimestampDuration = c.TimestampDuration()
				c.cacheDuration = c.Duration()
			}
			c.realTimeStart = time.Now()
			c.realTimeCount = 0
		}
	}

	diff := timestamp - c.lastTimestamp
	if diff > 0 {
		c.MaxGap = max(c.MaxGap, diff)
	} else if diff < 0 {
		c.MaxRewind = max(c.MaxRewind, -diff)
	} else {
		c.Duplicate++
	}
	c.Total++
	c.lastTimestamp = timestamp
}
