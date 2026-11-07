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
// 纯自动探测文件后缀对应注释风格；若无法识别则以 // 统一保底
func GetCommentedSeparator(filename string) (string, string, string) {
	separator := strings.Repeat("=", 80)

	// 1. 尝试自动探测后缀注释符
	cStart, cEnd := detectStyleByExt(filename)

	// 2. 若未识别后缀，以 // 作为通用保底注释
	if cStart == "" {
		cStart, cEnd = "//", ""
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
	// 阵营 1: 双斜杠 (//) - 现代语言主力军、底层计算、着色器及 IDL
	case ".c", ".h", ".cpp", ".hpp", ".cc", ".cxx", ".hh", ".hxx", ".inc", ".inl", ".ipp", ".tpp",
		".go", ".java", ".js", ".ts", ".jsx", ".tsx", ".rust", ".rs", ".php", ".cs", ".swift",
		".kt", ".kts", ".scala", ".dart", ".typ", ".scss", ".less",
		".zig", ".v", ".odin", ".mojo",
		".glsl", ".vert", ".frag", ".geom", ".comp", ".hlsl", ".metal", ".wgsl", ".cu", ".cuh", ".cl",
		".groovy", ".gvy", ".gradle", ".fs", ".fsi", ".fsx",
		".svelte", ".sol", ".proto", ".thrift", ".jsonc", ".json5":
		return "//", ""

	// 阵营 2: 井号 (#) - 脚本语言、IaC 基础设施及各类工程配置
	case ".py", ".sh", ".bash", ".zsh", ".fish", ".ksh", ".csh", ".tcsh", ".awk", ".sed", ".env",
		".yaml", ".yml", ".conf", ".ini", ".dockerfile", ".makefile", ".rb", ".pl", ".r", ".toml",
		".nim", ".jl", ".cr", ".ex", ".exs", ".tcl", ".raku",
		".cmake", ".properties", ".gitconfig", ".editorconfig", ".service", ".spec", ".tf", ".tfvars":
		return "#", ""

	// 阵营 3: 双横线 (--) - SQL 与特定函数式语言
	case ".sql", ".lua", ".hs", ".vhdl", ".vhd", ".ada",
		".elm", ".purs", ".agda", ".idr", ".applescript":
		return "--", ""

	// 阵营 4: HTML/XML 风格 (<!-- -->) - 标记语言与富文档
	case ".html", ".htm", ".xhtml", ".xml", ".md", ".qmd", ".markdown", ".mdx",
		".vue", ".astro", ".svg", ".xaml", ".plist", ".rss", ".atom":
		return "<!--", "-->"

	// 阵营 5: 分号 (;) - Lisp 族语言、汇编及 Windows 配置
	case ".lisp", ".lsp", ".el", ".scm", ".rkt", ".clj", ".cljs", ".cljc", ".edn",
		".asm", ".s", ".nasm", ".inf":
		return ";", ""

	// 阵营 6: 百分号 (%) - 科学计算与特定函数式语言
	case ".tex", ".latex", ".sty", ".cls", ".m", ".erl", ".hrl":
		return "%", ""

	// 阵营 7: 单引号 (') - BASIC 家族
	case ".vb", ".vba", ".vbs", ".bas":
		return "'", ""

	// 阵营 8: 块注释 (/* */)
	case ".css":
		return "/*", "*/"
	}

	return "", "" // 未识别后缀，由上层以 // 统一优雅保底
}
