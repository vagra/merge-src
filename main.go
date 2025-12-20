package main

import (
	"bufio"
	"fmt"
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
	if runtime.GOOS == "windows" {
		enableWindowsANSI()
	}

	// 1. 读取配置
	configPath := ".mergerule"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("%s错误: 未找到 .mergerule%s\n", ColorRed, ColorReset)
		os.Exit(1)
	}

	config := parseConfig(configPath)
	state := newState(config)

	// 2. 准备输出
	timestamp := time.Now().Format("20060102-1504")
	outputName := fmt.Sprintf("%s-%s.txt", config.OutputPrefix, timestamp)
	outFile, err := os.Create(outputName)
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	fmt.Println("--- 智能遍历目录 (Go Recursive) ---")

	// 3. 开始递归遍历 (从当前目录 ".")
	// 使用手动递归代替 filepath.WalkDir 以便控制检查时机
	processDir(".", config, state, outFile)

	// 4. 检查无效目录 (兜底检查那些连根都没找到的)
	// 注意：文件缺失已经在遍历过程中输出了，这里只报“路径写错了找不到目录”的情况
	hasInvalidDir := false
	for dir := range state.ExpectationsByDir {
		if dir == "" { continue }
		if !state.VisitedDirs[dir] {
			if !hasInvalidDir {
				fmt.Println("\n--- 检查无效路径 ---")
				hasInvalidDir = true
			}
			fmt.Printf("%s目录未找到: %s (其下指定文件均缺失)%s\n", ColorRed, filepath.FromSlash(dir), ColorReset)
		}
	}

	fmt.Printf("\n%s✔ 完成: %s%s\n", ColorGreen, outputName, ColorReset)
}

// --- 核心递归函数 ---
func processDir(dirPath string, cfg Config, state *AnalysisState, outFile *os.File) {
	// 1. 读取目录内容
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		// 如果读不到目录，可能是权限或不存在，跳过
		return
	}

	// 规范化路径
	relDir := filepath.ToSlash(dirPath)
	if relDir == "." { relDir = "" }
	
	state.VisitedDirs[relDir] = true

	// 2. 第一轮循环：仅处理文件 (合并)
	separator := strings.Repeat("=", 80)
	cStart, cEnd := getCommentStyle(cfg.CommentStyle)
	
	// [修改点] 构造带注释的分隔线
	var commentedSeparator string
	if cEnd == "" {
		// 单行注释风格 (如 // 或 #)
		commentedSeparator = fmt.Sprintf("%s %s", cStart, separator)
	} else {
		// 包裹注释风格 (如 <!-- -->)
		commentedSeparator = fmt.Sprintf("%s %s %s", cStart, separator, cEnd)
	}

	// 1. 处理文件
	for _, entry := range entries {
		if entry.IsDir() { continue } // 先跳过目录

		fileName := entry.Name()
		fullPath := filepath.Join(dirPath, fileName)
		relPath := filepath.ToSlash(fullPath) // 转为 / 分隔

		// 规则匹配
		action := matchRule(relPath, cfg.Rules)
		shouldInclude := (action == ActionInclude)

		if shouldInclude {
			// 后缀与显式检查
			ext := filepath.Ext(fileName)
			isExplicit := state.ExplicitFiles[relPath]
			isSource := false
			for _, e := range cfg.Extensions {
				if strings.EqualFold(ext, e) { isSource = true; break }
			}

			if !isExplicit && !isSource {
				shouldInclude = false
			}
		}

		if shouldInclude {
			fmt.Printf("%s合并: %s%s\n", ColorGreen, fullPath, ColorReset)

			// [修改点] 使用带注释的分隔线写入
			outFile.WriteString(fmt.Sprintf("\n%s\n", commentedSeparator))
			outFile.WriteString(fmt.Sprintf("%s FILE: %s %s\n", cStart, fullPath, cEnd))
			outFile.WriteString(fmt.Sprintf("%s\n", commentedSeparator))

			content, err := os.ReadFile(fullPath)
			if err == nil {
				outFile.Write(content)
				outFile.WriteString("\n")
			} else {
				outFile.WriteString(fmt.Sprintf("// Error reading file: %v\n", err))
			}

			state.MergedFiles[relPath] = true
		}
	}

	// 2. 检查缺失
	if expectedFiles, ok := state.ExpectationsByDir[relDir]; ok {
		for fname := range expectedFiles {
			// 构造期待的相对路径
			var expectedRel string
			if relDir == "" {
				expectedRel = fname
			} else {
				expectedRel = relDir + "/" + fname
			}

			if !state.MergedFiles[expectedRel] {
				// 使用系统分隔符打印，为了好看
				displayPath := filepath.FromSlash(expectedRel)
				fmt.Printf("%s缺失: %s%s\n", ColorRed, displayPath, ColorReset)
			}
		}
	}

	// 3. 递归子目录
	for _, entry := range entries {
		if !entry.IsDir() { continue }

		subDirName := entry.Name()
		subDirPath := filepath.Join(dirPath, subDirName)
		relSubPath := filepath.ToSlash(subDirPath)

		// 剪枝逻辑
		if shouldTraverse(relSubPath, cfg, state) {
			processDir(subDirPath, cfg, state, outFile)
		}
	}
}

// --- 辅助函数 ---

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
					if dir == "." { dir = "" }
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
	if err != nil { panic(err) }
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
		if line == "" || strings.HasPrefix(line, "#") { continue }
		
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
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			isInclude := strings.HasPrefix(line, "+")
			rawPath := strings.TrimSpace(line[1:])
			slashPath := filepath.ToSlash(rawPath)
			cleanPath := filepath.ToSlash(filepath.Clean(rawPath))
			isDir := strings.HasSuffix(slashPath, "/")
			isWild := strings.HasSuffix(slashPath, ".*")
			
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
			if relPath == r.Path || strings.HasPrefix(relPath, r.Path+"/") { matched = true }
		} else if r.IsWildcard {
			baseRule := strings.TrimSuffix(r.Path, ".*")
			if strings.HasPrefix(relPath, baseRule) {
				if len(relPath) > len(baseRule) && relPath[len(baseRule)] == '.' { matched = true }
			}
		} else {
			if relPath == r.Path { matched = true }
		}

		if matched {
			rLen := len(r.Path)
			if rLen > bestLen {
				bestLen = rLen
				if r.IsInclude { bestAction = ActionInclude } else { bestAction = ActionExclude }
			}
		}
	}
	return bestAction
}

func shouldTraverse(dirPath string, cfg Config, state *AnalysisState) bool {
	action := matchRule(dirPath, cfg.Rules)
	if action == ActionInclude { return true }
	if state.TraversePaths[dirPath] { return true }
	return false
}

func getCommentStyle(style string) (string, string) {
	switch strings.ToLower(style) {
	case "python", "ruby", "perl", "sh", "yaml", "conf", "ini", "dockerfile", "makefile": return "#", ""
	case "sql", "lua", "haskell": return "--", ""
	case "html", "xml": return "<!--", "-->"
	default: return "//", ""
	}
}

func enableWindowsANSI() {
	// Standard Windows 10/11 terminal support
}