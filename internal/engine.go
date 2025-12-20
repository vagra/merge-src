package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Engine 结构体
type Engine struct {
	Cfg               *Config // 直接使用 Config，无需 import
	ExplicitFiles     map[string]bool
	ExpectationsByDir map[string]map[string]bool
	MergedFiles       map[string]bool
	VisitedDirs       map[string]bool
	TraversePaths     map[string]bool
}

func NewEngine(cfg *Config) *Engine {
	e := &Engine{
		Cfg:               cfg,
		ExplicitFiles:     make(map[string]bool),
		ExpectationsByDir: make(map[string]map[string]bool),
		MergedFiles:       make(map[string]bool),
		VisitedDirs:       make(map[string]bool),
		TraversePaths:     make(map[string]bool),
	}
	e.initRules()
	return e
}

func (e *Engine) initRules() {
	for _, r := range e.Cfg.Rules {
		if r.IsInclude {
			parts := strings.Split(r.Path, "/")
			for i := 1; i < len(parts); i++ {
				e.TraversePaths[strings.Join(parts[:i], "/")] = true
			}
			if !r.IsDir {
				var targets []string
				if r.IsWildcard {
					base := strings.TrimSuffix(r.Path, ".*")
					for _, ext := range e.Cfg.Extensions {
						targets = append(targets, base+ext)
					}
				} else {
					targets = append(targets, r.Path)
				}
				for _, t := range targets {
					e.ExplicitFiles[t] = true
					dir := filepath.ToSlash(filepath.Dir(t))
					if dir == "." {
						dir = ""
					}
					if e.ExpectationsByDir[dir] == nil {
						e.ExpectationsByDir[dir] = make(map[string]bool)
					}
					e.ExpectationsByDir[dir][filepath.Base(t)] = true
				}
			}
		}
	}
}

func (e *Engine) Run(startDir string, outFile *os.File) {
	fmt.Println("--- 智能遍历目录 (Go Engine / Internal Pkg) ---")
	e.processDir(startDir, outFile)
	e.reportInvalidDirs()
}

func (e *Engine) processDir(dirPath string, outFile *os.File) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return
	}

	relDir := filepath.ToSlash(dirPath)
	if relDir == "." {
		relDir = ""
	}
	e.VisitedDirs[relDir] = true

	// 直接调用 util.go 里的函数
	sepLine, cStart, cEnd := GetCommentedSeparator(e.Cfg.CommentStyle)

	// 1. 文件处理
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		fullPath := filepath.Join(dirPath, fileName)
		relPath := filepath.ToSlash(fullPath)

		if e.shouldMerge(relPath, fileName) {
			fmt.Printf("%s合并: %s%s\n", ColorGreen, fullPath, ColorReset)

			outFile.WriteString(fmt.Sprintf("\n%s\n", sepLine))
			outFile.WriteString(fmt.Sprintf("%s FILE: %s %s\n", cStart, fullPath, cEnd))
			outFile.WriteString(fmt.Sprintf("%s\n", sepLine))

			content, err := os.ReadFile(fullPath)
			if err == nil {
				outFile.Write(content)
				outFile.WriteString("\n")
			} else {
				outFile.WriteString(fmt.Sprintf("// Error reading file: %v\n", err))
			}
			e.MergedFiles[relPath] = true
		}
	}

	// 2. 缺失检查
	if expectedFiles, ok := e.ExpectationsByDir[relDir]; ok {
		for fname := range expectedFiles {
			var expectedRel string
			if relDir == "" {
				expectedRel = fname
			} else {
				expectedRel = relDir + "/" + fname
			}
			if !e.MergedFiles[expectedRel] {
				fmt.Printf("%s缺失: %s%s\n", ColorRed, filepath.FromSlash(expectedRel), ColorReset)
			}
		}
	}

	// 3. 递归
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(dirPath, entry.Name())
		relSub := filepath.ToSlash(subDir)
		if e.shouldTraverse(relSub) {
			e.processDir(subDir, outFile)
		}
	}
}

func (e *Engine) shouldMerge(relPath string, fileName string) bool {
	action := e.matchRule(relPath)
	shouldInclude := (action == 1)
	if shouldInclude {
		ext := filepath.Ext(fileName)
		isExplicit := e.ExplicitFiles[relPath]
		isSource := false
		for _, extItem := range e.Cfg.Extensions {
			if strings.EqualFold(ext, extItem) {
				isSource = true
				break
			}
		}
		if !isExplicit && !isSource {
			shouldInclude = false
		}
	}
	return shouldInclude
}

func (e *Engine) shouldTraverse(relDir string) bool {
	return e.matchRule(relDir) == 1 || e.TraversePaths[relDir]
}

func (e *Engine) matchRule(relPath string) int {
	bestLen := -1
	bestAction := 0
	for _, r := range e.Cfg.Rules {
		matched := false
		if r.IsDir {
            // 增加对根目录 "." 的支持
			if r.Path == "." {
				matched = true
			} else if relPath == r.Path || strings.HasPrefix(relPath, r.Path+"/") {
				matched = true
			}
		} else if r.IsWildcard {
			base := strings.TrimSuffix(r.Path, ".*")
			if strings.HasPrefix(relPath, base) && len(relPath) > len(base) && relPath[len(base)] == '.' {
				matched = true
			}
		} else {
			if relPath == r.Path {
				matched = true
			}
		}
		if matched && len(r.Path) > bestLen {
			bestLen = len(r.Path)
			if r.IsInclude {
				bestAction = 1
			} else {
				bestAction = 2
			}
		}
	}
	return bestAction
}

func (e *Engine) reportInvalidDirs() {
	hasInvalid := false
	for dir := range e.ExpectationsByDir {
		if dir == "" {
			continue
		}
		if !e.VisitedDirs[dir] {
			if !hasInvalid {
				fmt.Println("\n--- 检查无效路径 ---")
				hasInvalid = true
			}
			fmt.Printf("%s目录未找到: %s (其下指定文件均缺失)%s\n", ColorRed, filepath.FromSlash(dir), ColorReset)
		}
	}
}
