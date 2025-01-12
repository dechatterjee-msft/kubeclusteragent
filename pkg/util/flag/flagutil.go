package flag

import (
	"flag"
	"os"
	"strconv"
)

// EnvStringVar sets a string flag from environment or defaults to a command line flag.
func EnvStringVar(s *string, envKey, name, value, usage string) {
	envVal := os.Getenv(envKey)
	if envVal != "" {
		value = envVal
	}

	flag.StringVar(s, name, value, usage)
}

// EnvBoolVar sets a bool flag from environment or defaults to a command line flag.
func EnvBoolVar(s *bool, envKey, name string, value bool, usage string) {
	var tf bool

	v, ok := os.LookupEnv(envKey)
	if ok {
		b, err := strconv.ParseBool(v)
		if err == nil {
			tf = b
		}
	}

	flag.BoolVar(s, name, tf, usage)
}
