package internal

import (
	"bufio"
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// Config 封装合并规则配置
type Config struct {
	Matcher gitignore.Matcher
}

// ParseConfig 解析 .mergerule 规则文件并构建标准 Gitignore 匹配器
func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []gitignore.Pattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 将规则解析为相对于项目根目录的标准 Pattern
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &Config{
		Matcher: gitignore.NewMatcher(patterns),
	}, nil
}
