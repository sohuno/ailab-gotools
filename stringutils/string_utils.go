package stringutils

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

func SplitString(line string) []string {
	return SplitStringWith(line, " ")
}

func SplitStringWith(line string, sep string) []string {
	var items []string
	for _, raw := range strings.Split(line, sep) {
		if raw = strings.Trim(raw, sep); len(raw) == 0 {
			continue
		}
		items = append(items, raw)
	}
	return items
}

func ListContains(list []string, target string) bool {
	for _, one := range list {
		if one == target {
			return true
		}
	}
	return false
}

func EnsureEndsWithSlash(str string) string {
	slashCount := 0
	if len(str) > 0 {
		for i := len(str) - 1; i >= 0; i-- {
			if '/' != str[i] {
				break
			}
			slashCount++
		}
	}
	if slashCount == 0 {
		return str + "/"
	} else if slashCount == 1 {
		return str
	} else {
		return str[:len(str)+1-slashCount]
	}
}

func IndexOfChar(str string, target byte, startPos int, nth int) int {
	if nth == 0 {
		return -1
	}
	count := nth
	for i := startPos; i < len(str); i++ {
		if target == str[i] {
			count--
			if count == 0 {
				return i
			}
		}
	}
	return -1
}

func StringToMd5(str string) string {
	md5Bytes := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", md5Bytes)
}

func SplitStringIgnoreInsideQuotation(str string) []string {
	str = strings.TrimSpace(str)
	re := regexp.MustCompile(`("[^"]+?"\S*|\S+)`)
	args := re.FindAllString(str, -1)
	return args
}

func Int64ArrayToString(arr []int64) string {
	var arrayString []string
	for _, value := range arr {
		arrayString = append(arrayString, strconv.FormatInt(value, 10))
	}
	result := strings.Join(arrayString, ",")
	return result
}

// ConvertRangeToInt64Array converts a string like '1-3,7' to an array of integers [1,2,3,7].
func ConvertRangeToInt64Array(rangeStr string) []int64 {
	var result []int64
	substrings := strings.Split(rangeStr, ",")

	for _, substr := range substrings {
		dashIndex := strings.Index(substr, "-")
		if dashIndex == -1 {
			num, err := strconv.ParseInt(substr, 10, 64)
			if err != nil {
				klog.Error("ParseInt64Fail. Str %s Err %v", substr, err)
				continue
			}
			result = append(result, num)
			continue
		}
		start, err := strconv.Atoi(substr[:dashIndex])
		end, err := strconv.Atoi(substr[dashIndex+1:])
		if err != nil {
			klog.Error("ParseInt64Fail. Str %s Err %v", substr, err)
			continue
		}

		for i := start; i <= end; i++ {
			result = append(result, int64(i))
		}
	}

	return result
}
