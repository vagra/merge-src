package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// --- 颜色定义 ---
var (
	ColorGreen = "\033[92m"
	ColorRed   = "\033[91m"
	ColorReset = "\033[0m"
)

// --- 配置结构 ---
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

// --- 状态追踪 ---
type AnalysisState struct {
	ExplicitFiles     map[string]bool
	ExpectationsByDir map[string]map[string]bool
	MergedFiles       map[string]bool
	VisitedDirs       map[string]bool
	TraversePaths     map[string]bool
}

func main() {
	// Windows 颜色兼容处理
	if runtime.GOOS == "windows" {
		enableWindowsANSI()
	}

	// 1. 读取配置文件
	configPath := ".mergerule"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("%s错误: 当前目录下未找到 .mergerule 配置文件%s\n", ColorRed, ColorReset)
		fmt.Println("请创建一个 .mergerule 文件并定义规则。")
		os.Exit(1)
	}

	config := parseConfig(configPath)
	state := newState(config)

	// 2. 准备输出文件
	timestamp := time.Now().Format("20060102-1504")
	outputName := fmt.Sprintf("%s-%s.txt", config.OutputPrefix, timestamp)
	outFile, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// 3. 开始遍历
	fmt.Println("--- 智能遍历目录 (Go Engine) ---")

	separator := strings.Repeat("=", 80)
	mergedCount := 0

	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath := filepath.ToSlash(path)
		if relPath == "." {
			return nil
		}

		// --- 目录处理 ---
		if d.IsDir() {
			state.VisitedDirs[relPath] = true
			if !shouldTraverse(relPath, config, state) {
				return filepath.SkipDir
			}
			return nil
		}

		// --- 文件处理 ---
		action := matchRule(relPath, config.Rules)
		shouldInclude := false

		if action == ActionInclude {
			shouldInclude = true
		} else if action == ActionExclude {
			shouldInclude = false
		}

		if shouldInclude {
			ext := filepath.Ext(d.Name())
			isExplicit := state.ExplicitFiles[relPath]
			isSource := false
			for _, e := range config.Extensions {
				if strings.EqualFold(ext, e) {
					isSource = true
					break
				}
			}

			if !isExplicit && !isSource {
				shouldInclude = false
			}
		}

		if shouldInclude {
			fmt.Printf("%s合并: %s%s\n", ColorGreen, path, ColorReset)

			cStart, cEnd := getCommentStyle(config.CommentStyle)
			outFile.WriteString(fmt.Sprintf("\n%s\n", separator))
			outFile.WriteString(fmt.Sprintf("%s FILE: %s %s\n", cStart, path, cEnd))
			outFile.WriteString(fmt.Sprintf("%s\n", separator))

			content, err := os.ReadFile(path)
			if err == nil {
				outFile.Write(content)
				outFile.WriteString("\n")
			} else {
				outFile.WriteString(fmt.Sprintf("// Error reading file: %v\n", err))
			}

			state.MergedFiles[relPath] = true
			mergedCount++
		}

		return nil
	})

	if err != nil {
		fmt.Printf("遍历出错: %v\n", err)
	}

	// 4. 完整性检查
	fmt.Println("\n--- 完整性检查 ---")
	missingCount := 0
	for dir, files := range state.ExpectationsByDir {
		for fname := range files {
			fullPath := dir
			if fullPath != "" {
				fullPath = fullPath + "/" + fname
			} else {
				fullPath = fname
			}

			if !state.MergedFiles[fullPath] {
				// 只有当目录确实被访问过（存在），或者目录虽然没访问过但我们也没报过目录丢失时
				if !state.VisitedDirs[dir] && dir != "" {
					continue
				}
				fmt.Printf("%s缺失: %s%s\n", ColorRed, filepath.FromSlash(fullPath), ColorReset)
				missingCount++
			}
		}
	}

	// 5. 无效目录检查
	fmt.Println("\n--- 检查无效目录 ---")
	for dir := range state.ExpectationsByDir {
		if dir == "" {
			continue
		}
		if !state.VisitedDirs[dir] {
			fmt.Printf("%s目录未找到: %s (其下指定文件均缺失)%s\n", ColorRed, filepath.FromSlash(dir), ColorReset)
		}
	}

	fmt.Printf("\n%s✔ 完成: %s (共 %d 个文件)%s\n", ColorGreen, outputName, mergedCount, ColorReset)
}

// --- 核心逻辑 ---

const (
	ActionNone    = 0
	ActionInclude = 1
	ActionExclude = 2
)

func newState(cfg Config) *AnalysisState {
	s := &AnalysisState{
		ExplicitFiles:     make(map[string]bool),
		ExpectationsByDir: make(map[string]map[string]bool),
		MergedFiles:       make(map[string]bool),
		VisitedDirs:       make(map[string]bool),
		TraversePaths:     make(map[string]bool),
	}

	for _, r := range cfg.Rules {
		if r.IsInclude {
			parts := strings.Split(r.Path, "/")
			for i := 1; i < len(parts); i++ {
				parent := strings.Join(parts[:i], "/")
				s.TraversePaths[parent] = true
			}
			if !r.IsDir {
				var targets []string
				if r.IsWildcard {
					base := strings.TrimSuffix(r.Path, ".*")
					for _, ext := range cfg.Extensions {
						targets = append(targets, base+ext)
					}
				} else {
					targets = append(targets, r.Path)
				}
				for _, t := range targets {
					s.ExplicitFiles[t] = true
					dir := filepath.ToSlash(filepath.Dir(t))
					if dir == "." {
						dir = ""
					}
					if s.ExpectationsByDir[dir] == nil {
						s.ExpectationsByDir[dir] = make(map[string]bool)
					}
					s.ExpectationsByDir[dir][filepath.Base(t)] = true
				}
			}
		}
	}
	return s
}

func parseConfig(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	cfg := Config{
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
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			isInclude := strings.HasPrefix(line, "+")
			rawPath := strings.TrimSpace(line[1:])
			slashPath := filepath.ToSlash(rawPath)
			cleanPath := filepath.ToSlash(filepath.Clean(rawPath)) // Clean removes trailing slash

			// Hack: Clean removes trailing slash, check original slashPath for Dir/Wildcard info
			isDir := strings.HasSuffix(slashPath, "/")
			isWild := strings.HasSuffix(slashPath, ".*")

			// 如果是目录，Clean 后的路径必须被用来做前缀匹配

			cfg.Rules = append(cfg.Rules, Rule{
				Path:       cleanPath,
				IsInclude:  isInclude,
				IsDir:      isDir,
				IsWildcard: isWild,
			})
		}
	}
	return cfg
}

func matchRule(relPath string, rules []Rule) int {
	bestLen := -1
	bestAction := ActionNone

	for _, r := range rules {
		matched := false
		if r.IsDir {
			if relPath == r.Path || strings.HasPrefix(relPath, r.Path+"/") {
				matched = true
			}
		} else if r.IsWildcard {
			baseRule := strings.TrimSuffix(r.Path, ".*")
			if strings.HasPrefix(relPath, baseRule) {
				if len(relPath) > len(baseRule) && relPath[len(baseRule)] == '.' {
					matched = true
				}
			}
		} else {
			if relPath == r.Path {
				matched = true
			}
		}

		if matched {
			rLen := len(r.Path)
			if rLen > bestLen {
				bestLen = rLen
				if r.IsInclude {
					bestAction = ActionInclude
				} else {
					bestAction = ActionExclude
				}
			}
		}
	}
	return bestAction
}

func shouldTraverse(dirPath string, cfg Config, state *AnalysisState) bool {
	action := matchRule(dirPath, cfg.Rules)
	if action == ActionInclude {
		return true
	}
	if state.TraversePaths[dirPath] {
		return true
	}
	return false
}

func getCommentStyle(style string) (string, string) {
	switch strings.ToLower(style) {
	case "python", "ruby", "perl", "sh", "yaml", "conf", "ini", "dockerfile":
		return "#", ""
	case "sql", "lua", "haskell":
		return "--", ""
	case "html", "xml":
		return "<!--", "-->"
	case "c", "cpp", "go", "java", "js", "ts", "php", "rust", "cs", "swift":
		return "//", ""
	default:
		return "//", ""
	}
}

func enableWindowsANSI() {
	// 这是一个最简易的实现，确保 Windows CMD/PowerShell 能显示颜色
	// 实际上 Go 的标准库在 Windows 上默认不处理 ANSI 转义。
	// 为了不引入第三方库，我们只做一个空调用，依赖现代终端 (Windows Terminal, Git Bash, VSCode) 自带的支持。
	// 现代 Windows 10/11 的终端默认都支持。
	// 如果必须支持旧版 CMD，需要 syscall 调用 SetConsoleMode，这里为了单文件简洁性省略。
}
