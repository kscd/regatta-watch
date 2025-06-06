package main

import "time"

type timeSource interface {
	Now() time.Time
	Since(t time.Time) time.Duration
}

type realTimeSource struct{}

func (r *realTimeSource) Now() time.Time {
	return time.Now()
}

func (r *realTimeSource) Since(t time.Time) time.Duration {
	return time.Since(t)
}

type clock struct {
	timeSource        timeSource
	referenceTime     time.Time // Base time for calculations
	referenceRealTime time.Time // When the reference was set in real time
	speed             float64   // Speed factor (1.0 is normal speed)
}

// newClock creates a new clock with default settings (no offset, normal speed)
func newClock() *clock {
	now := time.Now()
	return &clock{
		timeSource:        &realTimeSource{},
		referenceTime:     now,
		referenceRealTime: now,
		speed:             1.0,
	}
}

// Now returns the current time adjusted by offset and speed
func (c *clock) Now() time.Time {
	elapsed := c.timeSource.Since(c.referenceRealTime)
	adjustedElapsed := time.Duration(float64(elapsed) * c.speed)
	return c.referenceTime.Add(adjustedElapsed)
}

// RealNow Always returns the actual real time
func (c *clock) RealNow() time.Time {
	return c.timeSource.Now()
}

// SetSpeed sets the time speed factor
func (c *clock) SetSpeed(speed float64) {
	now := c.timeSource.Now()
	adjustedNow := c.Now()

	c.referenceTime = adjustedNow
	c.referenceRealTime = now
	c.speed = speed
}

// SetOffset sets the time offset
func (c *clock) SetOffset(offset time.Duration) {
	now := c.timeSource.Now()

	c.referenceTime = now.Add(-offset)
	c.referenceRealTime = now
}

// SetTimeMapping sets the clock to report fakeTime when the real time is realTime
func (c *clock) SetTimeMapping(realTime, fakeTime time.Time) {
	c.referenceTime = fakeTime
	c.referenceRealTime = realTime
}

// SetCurrentTimeAs sets the clock to report the specified fakeTime at the current moment
func (c *clock) SetCurrentTimeAs(fakeTime time.Time) {
	realTime := c.timeSource.Now()

	c.referenceTime = fakeTime
	c.referenceRealTime = realTime
}

// Reset resets the clock to current time with no offset and normal speed
func (c *clock) Reset() {
	now := c.timeSource.Now()
	c.referenceTime = now
	c.referenceRealTime = now
	c.speed = 1.0
}
