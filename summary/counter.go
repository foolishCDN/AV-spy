package summary

import (
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
	realTimeFps            int
	cacheTimestampDuration int
	cacheDuration          time.Duration
	DiffThreshold          int
}

func (c *Counter) TimestampDuration() int {
	return c.lastTimestamp - c.firstTimestamp
}

func (c *Counter) Duration() time.Duration {
	return time.Since(c.startTime)
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
		if time.Since(c.realTimeStart) >= time.Second {
			c.realTimeFps = c.realTimeCount / int(time.Since(c.realTimeStart))
			if (float64(c.realTimeFps)-c.Rate())/c.Rate()*100 < float64(c.DiffThreshold) {
				// dry of server cache
				c.cacheTimestampDuration = c.TimestampDuration()
				c.cacheDuration = c.Duration()
			}
			c.realTimeStart = time.Now()
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
