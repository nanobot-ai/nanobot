package expr

import "strings"

func Lookup(envMap map[string]string, envKey string) (string, bool) {
	val, ok := envMap[envKey]
	if ok {
		return val, true
	}
	for envMapKey, envMapVal := range envMap {
		if strings.EqualFold(envKey, strings.ReplaceAll(envMapKey, "-", "_")) {
			val = envMapVal
			ok = true
			break
		}
	}
	if ok {
		return val, true
	}

	return "", false
}
