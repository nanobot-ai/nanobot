package envvar

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/obot-platform/nanobot/pkg/expr"
	"github.com/obot-platform/nanobot/pkg/log"
)

func ReplaceString(envs map[string]string, str string) string {
	r, err := expr.EvalString(context.TODO(), envs, str)
	if err != nil {
		log.Errorf(context.TODO(), "failed to evaluate expression %s: %v", str, err)
		return str
	}
	return r
}

func ReplaceMap(envs map[string]string, m map[string]string) map[string]string {
	newMap := make(map[string]string, len(m))
	for k, v := range m {
		newMap[ReplaceString(envs, k)] = ReplaceString(envs, v)
	}
	return newMap
}

func ReplaceEnv(envs map[string]string, command string, args []string, env map[string]string) (string, []string, []string) {
	newEnvMap := make(map[string]string, len(envs)+len(env))
	for k, v := range envs {
		if !strings.Contains(k, ":") {
			newEnvMap[k] = v
		}
	}

	maps.Copy(newEnvMap, ReplaceMap(envs, env))

	newEnv := make([]string, 0, len(env))
	for _, k := range slices.Sorted(maps.Keys(newEnvMap)) {
		newEnv = append(newEnv, fmt.Sprintf("%s=%s", k, newEnvMap[k]))
	}

	newArgs := make([]string, len(args))
	for i, arg := range args {
		newArgs[i] = ReplaceString(envs, arg)
	}
	return ReplaceString(envs, command), newArgs, newEnv
}
