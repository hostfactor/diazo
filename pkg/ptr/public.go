package ptr

import "time"

func String(v string) *string {
	return &v
}

func Uint(v uint) *uint {
	return &v
}

func Uint8(v uint8) *uint8 {
	return &v
}

func Uint16(v uint16) *uint16 {
	return &v
}

func Uint32(v uint32) *uint32 {
	return &v
}

func Uint64(v uint64) *uint64 {
	return &v
}

func Int(v int) *int {
	return &v
}

func Int8(v int8) *int8 {
	return &v
}

func Int16(v int16) *int16 {
	return &v
}

func Int32(v int32) *int32 {
	return &v
}

func Int64(v int64) *int64 {
	return &v
}

func Float64(v float64) *float64 {
	return &v
}

func Float32(v float32) *float32 {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

func Duration(v time.Duration) *time.Duration {
	return &v
}

func DerefDuration(v *time.Duration) time.Duration {
	if v == nil {
		return 0
	}
	return *v
}

func DerefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func DerefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func DerefUint32(v *uint32) uint32 {
	if v == nil {
		return 0
	}
	return *v
}

func DerefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
