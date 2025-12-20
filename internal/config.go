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
	Path       string
	IsInclude  bool
	IsDir      bool
	IsWildcard bool
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
					if !strings.HasPrefix(e, ".") {
						e = "." + e
					}
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

		// Rules
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			isInclude := strings.HasPrefix(line, "+")
			rawPath := strings.TrimSpace(line[1:])
			slashPath := filepath.ToSlash(rawPath)
			cleanPath := filepath.ToSlash(filepath.Clean(rawPath))

			cfg.Rules = append(cfg.Rules, Rule{
				Path:       cleanPath,
				IsInclude:  isInclude,
				IsDir:      strings.HasSuffix(slashPath, "/"),
				IsWildcard: strings.HasSuffix(slashPath, ".*"),
			})
		}
	}
	return cfg, nil
}
