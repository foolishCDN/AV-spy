package summary

import (
	"math"
	"time"

	"github.com/sirupsen/logrus"
)

func NewCounter(opts ...CounterOption) *Counter {
	c := &Counter{
		HintGap:  200,
		HintHole: 200 * time.Millisecond,
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

type Counter struct {
	LogPrefix string
	Total     int
	MaxGap    int
	MaxRewind int
	Duplicate int
	MaxHole   time.Duration

	HintGap  int
	HintHole time.Duration

	firstTimestamp  int
	lastTimestamp   int
	lastReceiveTime time.Time

	// for computing the cache content of live stream in server
	startTime              time.Time
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
		c.realTimeCount++

		c.firstTimestamp = timestamp
		c.lastTimestamp = timestamp
		c.lastReceiveTime = now
		c.Total++
		return
	}

	// compute real time fps
	if c.cacheTimestampDuration == 0 {
		c.realTimeCount++
		if time.Since(c.realTimeStart) >= time.Millisecond*100 {
			c.estimatedCacheFps = float64(c.realTimeCount) / time.Since(c.realTimeStart).Seconds()
			if math.Abs(float64(c.estimatedCacheFps)-c.Rate())/c.Rate()*100 < float64(c.DiffThreshold) {
				// it is considered that all the server cache has been consumed
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
		if c.MaxGap > c.HintGap {
			logrus.WithFields(logrus.Fields{
				"gap":  diff,
				"max":  c.MaxGap,
				"last": c.lastTimestamp,
				"now":  timestamp,
			}).Warnf("%s: dts jump", c.LogPrefix)
		}
	} else if diff < 0 {
		c.MaxRewind = max(c.MaxRewind, -diff)
		logrus.WithFields(logrus.Fields{
			"rewind": diff,
			"max":    c.MaxRewind,
			"last":   c.lastTimestamp,
			"now":    timestamp,
		}).Warnf("%s: dts rewind", c.LogPrefix)
	} else {
		c.Duplicate++
		logrus.WithFields(logrus.Fields{
			"last": c.lastTimestamp,
			"now":  timestamp,
		}).Warnf("%s: dts duplicate", c.LogPrefix)
	}
	hole := time.Since(c.lastReceiveTime)
	if hole > c.HintHole {
		c.MaxHole = max(c.MaxHole, hole)
		logrus.WithFields(logrus.Fields{
			"hole": hole.Milliseconds(),
			"max":  c.MaxHole.Milliseconds(),
			"last": c.lastReceiveTime.Format(time.RFC3339Nano),
			"now":  time.Now().Format(time.RFC3339Nano),
		}).Warnf("%s: data has hole", c.LogPrefix)
	}
	c.Total++
	c.lastTimestamp = timestamp
	c.lastReceiveTime = time.Now()
}
