package internal

import (
	"fmt"
	"strings"
)

// --- 颜色定义 ---
const (
	ColorGreen = "\033[92m"
	ColorRed   = "\033[91m"
	ColorReset = "\033[0m"
)

// GetCommentedSeparator 生成带注释的分隔线
func GetCommentedSeparator(style string) (string, string, string) {
	separator := strings.Repeat("=", 80)
	cStart, cEnd := getCommentStyle(style)

	var commentedSeparator string
	if cEnd == "" {
		commentedSeparator = fmt.Sprintf("%s %s", cStart, separator)
	} else {
		commentedSeparator = fmt.Sprintf("%s %s %s", cStart, separator, cEnd)
	}
	return commentedSeparator, cStart, cEnd
}

func getCommentStyle(style string) (string, string) {
	switch strings.ToLower(style) {
	case "python", "ruby", "perl", "sh", "yaml", "conf", "ini", "dockerfile", "makefile":
		return "#", ""
	case "sql", "lua", "haskell":
		return "--", ""
	case "html", "xml":
		return "<!--", "-->"
	default:
		return "//", ""
	}
}
