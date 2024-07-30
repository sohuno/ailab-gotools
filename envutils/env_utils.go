package sys_utils

import (
	"github.com/sohuno/gotools/stringutils"
	"os"
)

type SystemEnvironment struct {
	envFile string
	envMap  map[string]string
}

func NewSystemEnvironment() *SystemEnvironment {
	return &SystemEnvironment{
		envFile: "/proc/1/environ",
		envMap:  make(map[string]string, 0),
	}
}

func (m *SystemEnvironment) Load() error {
	buf, err := os.ReadFile(m.envFile)
	if err != nil {
		return err
	}
	var line string
	var start = 0
	for {
		line, start = m.parseOne(buf, start)
		if start == -1 {
			break
		}
		if len(line) == 0 {
			continue
		}
		items := stringutils.SplitStringWith(line, "=")
		if len(items) != 2 {
			continue
		}
		m.envMap[items[0]] = items[1]
	}
	return nil
}

func (m *SystemEnvironment) parseOne(buf []byte, i int) (string, int) {
	length := len(buf)
	if i >= length {
		return "", -1
	}
	if buf[i] == 0 {
		return "", i + 1
	}
	for j := i + 1; j < length; j++ {
		if buf[j] == 0 {
			return string(buf[i:j]), j + 1
		}
	}
	return string(buf[i:]), -1
}

func (m *SystemEnvironment) GetEnv(envName string) (string, bool) {
	val, found := m.envMap[envName]
	if !found {
		return "", found
	}
	return val, true
}
