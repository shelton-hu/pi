package redis

const (
	ExpireTimeSceond = 1
	ExpireTImeMinute = 60 * ExpireTimeSceond
	ExpireTimeHour   = 60 * ExpireTImeMinute
	ExpireTimeDay    = 24 * ExpireTimeHour
	ExpireTime30Day  = 30 * ExpireTimeDay
)
