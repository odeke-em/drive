package expirable

import "time"

type ExpirableValue struct {
	value     interface{}
	entryTime time.Time
}

func (e *ExpirableValue) Value() interface{} {
	if e == nil {
		return nil
	}
	return e.value
}

func (e *ExpirableValue) Expired(q time.Time) bool {
	if e == nil {
		return true
	}

	return e.entryTime.Before(q)
}

func NewExpirableValue(v interface{}) *ExpirableValue {
	return NewExpirableValueWithOffset(v, 0)
}

func NewExpirableValueWithOffset(v interface{}, expiry uint64) *ExpirableValue {
	return &ExpirableValue{
		value:     v,
		entryTime: time.Now().Add(time.Duration(expiry)),
	}
}
