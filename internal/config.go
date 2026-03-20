package internal

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Extensions   []string
	CommentStyle string
	OutputPrefix string
	Rules        []Rule
}

type Rule struct {
	Raw         string // 原始字符串，用于比较优先级长度
	IsInclude   bool
	
	BaseDir     string // 基础目录
	Recursive   bool   // 是否允许递归子目录
	Pattern     string // 文件名匹配模式 (Glob)
	CheckExts   bool   // 是否需要检查全局 extensions
}

// ParseConfig 解析配置文件
func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &Config{
		Extensions:   []string{".c", ".h"},
		CommentStyle: "c",
		OutputPrefix: "merge_src_result",
		Rules:        []Rule{},
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Key-Value
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "-") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(strings.ToLower(parts[0]))
			val := strings.TrimSpace(parts[1])
			switch key {
			case "extensions":
				exts := strings.Split(val, ",")
				var cleanExts []string
				for _, e := range exts {
					e = strings.TrimSpace(e)
					if !strings.HasPrefix(e, ".") { e = "." + e }
					cleanExts = append(cleanExts, e)
				}
				cfg.Extensions = cleanExts
			case "style":
				cfg.CommentStyle = val
			case "output_prefix":
				cfg.OutputPrefix = val
			}
			continue
		}

		// Rules Parser
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			isInclude := strings.HasPrefix(line, "+")
			rawPath := strings.TrimSpace(line[1:])
			// 统一使用 / 作为分隔符进行处理
			slashPath := filepath.ToSlash(rawPath)

			r := Rule{
				Raw:       slashPath,
				IsInclude: isInclude,
			}

			if strings.HasSuffix(slashPath, "/") {
				// 情况 1: 目录规则 (+path/)
				// [修复] filepath.Clean 在 Windows 会变回 \，所以外层必须套 filepath.ToSlash
				r.BaseDir = filepath.ToSlash(filepath.Clean(slashPath))
				r.Recursive = true
				r.Pattern = "" 
				r.CheckExts = true

			} else if strings.Contains(slashPath, "/**/") {
				// 情况 2: 递归通配规则 (+path/**/pattern)
				// 含义: 递归包含 path 下符合 pattern 的文件 (忽略全局 extensions)
				parts := strings.Split(slashPath, "/**/")
				// [修复] 强制转 /
				r.BaseDir = filepath.ToSlash(filepath.Clean(parts[0]))
				r.Recursive = true
				r.Pattern = parts[1] // 例如 *.h 或 *
				r.CheckExts = false

			} else if !strings.Contains(slashPath, "/") {
				// 情况 3: 全局匹配规则 (+pattern 或 -pattern)
				// 如果完全没有斜杠，则视为针对所有目录的递归匹配
				r.BaseDir = ""
				r.Pattern = slashPath
				r.Recursive = true
				r.CheckExts = false

			} else {
				// 情况 4: 平铺通配规则 (+path/pattern) 或 精确文件
				// [修复] filepath.Dir 在 Windows 会返回 \，强制转 /
				r.BaseDir = filepath.ToSlash(filepath.Dir(slashPath))
				r.Pattern = filepath.Base(slashPath)
				r.Recursive = false
				r.CheckExts = false
				
				// 根目录修正: 如果用户写了 +./foo，Dir("./foo") 返回 "."
				if r.BaseDir == "." && strings.HasPrefix(slashPath, "./") {
					r.BaseDir = ""
				}
			}

			// 再次确保 BaseDir 没有 . (Clean 可能会留下 .)
			if r.BaseDir == "." {
				r.BaseDir = ""
			}

			cfg.Rules = append(cfg.Rules, r)
		}
	}
	return cfg, nil
}
