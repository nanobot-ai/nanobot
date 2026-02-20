//go:build !linux

package system

import (
	"os"
	"reflect"
	"time"
)

func fileCreatedAt(_ string, info os.FileInfo) (time.Time, bool) {
	return createdAtFromSys(info.Sys())
}

func createdAtFromSys(sys any) (time.Time, bool) {
	if sys == nil {
		return time.Time{}, false
	}

	v := reflect.Indirect(reflect.ValueOf(sys))
	if v.Kind() != reflect.Struct {
		return time.Time{}, false
	}

	if nsec, ok := nanosecondsFromCreationTime(v); ok && nsec != 0 {
		return time.Unix(0, nsec), true
	}
	if t, ok := timeFromTimespecField(v, "Birthtimespec"); ok {
		return t, true
	}
	if t, ok := timeFromTimespecField(v, "Birthtim"); ok {
		return t, true
	}
	if t, ok := timeFromTimespecField(v, "X__st_birthtim"); ok {
		return t, true
	}

	return time.Time{}, false
}

func nanosecondsFromCreationTime(v reflect.Value) (int64, bool) {
	field := v.FieldByName("CreationTime")
	if !field.IsValid() {
		return 0, false
	}

	m := field.MethodByName("Nanoseconds")
	if !m.IsValid() && field.CanAddr() {
		m = field.Addr().MethodByName("Nanoseconds")
	}
	if !m.IsValid() || m.Type().NumIn() != 0 || m.Type().NumOut() != 1 {
		return 0, false
	}

	result := m.Call(nil)[0]
	switch result.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return result.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(result.Uint()), true
	default:
		return 0, false
	}
}

func timeFromTimespecField(v reflect.Value, fieldName string) (time.Time, bool) {
	ts := v.FieldByName(fieldName)
	if !ts.IsValid() {
		return time.Time{}, false
	}

	ts = reflect.Indirect(ts)
	if ts.Kind() != reflect.Struct {
		return time.Time{}, false
	}

	sec, secOK := intField(ts, "Sec")
	nsec, nsecOK := intField(ts, "Nsec")
	if !secOK || !nsecOK {
		sec, secOK = intField(ts, "Tv_sec")
		nsec, nsecOK = intField(ts, "Tv_nsec")
	}
	if !secOK || !nsecOK {
		return time.Time{}, false
	}
	if sec == 0 && nsec == 0 {
		return time.Time{}, false
	}

	return time.Unix(sec, nsec), true
}

func intField(v reflect.Value, fieldName string) (int64, bool) {
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return 0, false
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(field.Uint()), true
	default:
		return 0, false
	}
}
