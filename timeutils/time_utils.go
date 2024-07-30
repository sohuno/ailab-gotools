package timeutils

/*
 * ## conversion between time object and int64
 * ## time -> int64
 * now := time.Now()
 * ts := now.UnixNano()
 *
 * ## int64 -> time
 * now2 := time.Unix(0, ts)
 *
 */

import "time"

type TimeUtils interface {
	NowInMs() int64
	NowInSec() int64
}

type TimeUtilsImpl struct {
}

func (m *TimeUtilsImpl) NowInMs() int64 {
	return NowInMs()
}

func (m *TimeUtilsImpl) NowInSec() int64 {
	return NowInSec()
}

type TimeUtilsCustom struct {
	TimeInMs  int64
	TimeInSec int64
}

func (m *TimeUtilsCustom) NowInMs() int64 {
	return m.TimeInMs
}

func (m *TimeUtilsCustom) NowInSec() int64 {
	return m.TimeInSec
}

func NowInMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func NowInSec() int64 {
	return time.Now().UnixNano() / 1e9
}

func Format(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.999999999 -0700 MST")
}

// return timeSec, timeMs, timeReadable
func GetSlsLogTime() (int64, int64, string) {
	ts := time.Now()
	timeMs := ts.UnixNano() / 1e6
	timeSec := ts.UnixNano() / 1e9

	TimeFormat := "2006-01-02 15:04:05.000000"
	timeReadable := ts.Format(TimeFormat)
	return timeSec, timeMs, timeReadable
}

func FormatMilli(timeMs int64) string {
	t := time.UnixMilli(timeMs)
	return t.Format("2006-01-02 15:04:05")
}
