package internal

import (
	"fmt"
	"path/filepath"
	"strings"
)

// --- 颜色定义 ---
const (
	ColorGreen = "\033[92m"
	ColorRed   = "\033[91m"
	ColorReset = "\033[0m"
)

// FormatByteSize 将字节大小转换为对齐的字符串 (如 "  1.5 KB")
// 保持 9 个字符的固定宽度以确保控制台输出对齐
func FormatByteSize(s int64) string {
	f := float64(s)
	if s < 1024 {
		return fmt.Sprintf("%6d  B", s)
	} else if s < 1024*1024 {
		return fmt.Sprintf("%6.1f KB", f/1024.0)
	} else {
		return fmt.Sprintf("%6.1f MB", f/(1024.0*1024.0))
	}
}

// GetCommentedSeparator 生成带注释的分隔线
// 增加 filename 参数，用于自动识别
// fallbackStyle 用于当无法识别后缀时的默认风格(即配置文件里的 style)
func GetCommentedSeparator(filename string, fallbackStyle string) (string, string, string) {
	separator := strings.Repeat("=", 80)

	// 1. 尝试自动探测
	cStart, cEnd := detectStyleByExt(filename)

	// 2. 如果没探测到，使用默认配置
	if cStart == "" {
		cStart, cEnd = getCommentStyle(fallbackStyle)
	}

	var commentedSeparator string
	if cEnd == "" {
		commentedSeparator = fmt.Sprintf("%s %s", cStart, separator)
	} else {
		commentedSeparator = fmt.Sprintf("%s %s %s", cStart, separator, cEnd)
	}
	return commentedSeparator, cStart, cEnd
}

// detectStyleByExt 根据后缀返回注释符号
func detectStyleByExt(filename string) (string, string) {
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	// Group 1: Hash Style (#)
	case ".py", ".sh", ".bash", ".zsh", ".yaml", ".yml", ".conf", ".ini",
	     ".dockerfile", ".makefile", ".rb", ".pl", ".r", ".toml":
		return "#", ""

	// Group 2: Double Dash (--)
	case ".sql", ".lua", ".hs", ".vhdl", ".ada":
		return "--", ""

	// Group 3: HTML Style (<!-- -->)
	case ".html", ".xml", ".md", ".qmd", ".markdown", ".vue":
		return "<!--", "-->"

	// Group 4: Percent (%)
	case ".tex", ".latex", ".m": // Matlab/Octave/LaTeX
		return "%", ""

	// Group 5: C Style (//) - 大多数现代语言
	// 包括 .typ (Typst)
	case ".c", ".h", ".cpp", ".hpp", ".cc", ".go", ".java", ".js", ".ts",
	     ".jsx", ".tsx", ".rust", ".rs", ".php", ".cs", ".swift", ".kt",
	     ".scala", ".dart", ".typ", ".scss", ".less":
		return "//", ""

	// CSS (/* */) - 只有块注释
	case ".css":
		return "/*", "*/"
	}

	return "", "" // 未知，交由 fallback 处理
}

// getCommentStyle 解析配置文件中的字符串设置 (保持兼容)
func getCommentStyle(style string) (string, string) {
	switch strings.ToLower(style) {
	case "python", "ruby", "perl", "sh", "yaml", "conf", "ini", "dockerfile", "makefile", "shell":
		return "#", ""
	case "sql", "lua", "haskell":
		return "--", ""
	case "html", "xml", "markdown":
		return "<!--", "-->"
	case "tex", "latex":
		return "%", ""
	default:
		return "//", ""
	}
}
