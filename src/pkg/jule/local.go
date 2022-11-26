package jule

import (
	"fmt"
	"strings"
)

// DecodeLocalization decodes localization configuration text
// and sets destination map by INI content.
func DecodeLocalization(ini string, dest *map[string]string) error {
	switch {
	case ini == "":
		return nil
	case dest == nil:
		return fmt.Errorf("dest is nil")
	}

	lines := strings.SplitN(ini, "\n", -1)
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pair := strings.SplitN(line, " ", 2)
		key := pair[0]
		if len(pair) == 1 {
			return fmt.Errorf(`%d: missing key value: "%s"`, i+1, pair[0])
		}
		value := strings.TrimSpace(pair[1])
		_, ok := (*dest)[key]
		if !ok {
			return fmt.Errorf(`%d: invalid key: "%s"`, i+1, key)
		}
		(*dest)[key] = value
	}

	return nil
}
