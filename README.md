# merge-src: AI Code Context Builder

![License](https://img.shields.io/badge/license-MIT-blue.svg) ![Go](https://img.shields.io/badge/go-1.24+-00ADD8.svg)

> **Navigation**: [🇺🇸 English](#english-documentation) | [🇨🇳 中文说明](#chinese-documentation)

<div id="english-documentation"></div>

## 🇺🇸 English Documentation

**merge-src** is a lightweight, zero-config CLI tool designed to pack your project source code into a single, clean text file — perfectly tailored as **Code Context** for LLMs (Large Language Models) in browser chat boxes or prompt workflows.

Powered by standard `.gitignore` syntax, it allows developers to precisely curate which source files to include or exclude without any learning curve.

### ✨ Key Features

- **🎯 Standard Gitignore Engine**: Powered by `go-git`, supporting sequential evaluation, `!` re-inclusion, wildcards (`*`, `**`), and directory pruning. Zero learning cost.
- **⚙️ Zero-Config Design**: No tedious `key = value` options. Automatically names output files after the current project root directory (`<project>-YYYYMMDD-HHMM.txt`).
- **🌐 Seamless UTF-8 Transcoding**: Automatically detects and lossless converts legacy encodings (ANSI, GBK, GB2312, GB18030) and strips BOM headers, eliminating garbled Chinese comments in AI chats.
- **🎨 Auto Comment Delimiters**: Auto-detects 8 language families (C/C++, Scripting, SQL, Lisp/ASM, HTML, LaTeX, BASIC, CSS) to create clean, syntax-friendly file separators.
- **🛡️ Built-in Protection**: Automatically ignores `.git/`, `.mergerule`, `.gitignore`, and the generated output file to prevent recursive self-merging and metadata pollution.

### 📦 Installation

#### Option 1: Via Go (Recommended)

```bash
go install github.com/vagra/merge-src@latest
```

#### Option 2: Build from Source

```bash
git clone https://github.com/vagra/merge-src.git
cd merge-src
go build
# Output: merge-src (Linux/Mac) or merge-src.exe (Windows)
```

### 🚀 Quick Start

1. Go to your project root.
2. Create a rule file named `.mergerule`.
3. Run:

```bash
merge-src
```

### 📝 Rule Syntax (.mergerule)

`.mergerule` uses standard Gitignore syntax. Sequential evaluation applies: **later rules override earlier ones (Last Match Wins)**.

#### Pattern A: Blacklist Mode (Default: Merge All, Exclude Unwanted)

```gitignore
# Exclude build artifacts and dependencies
*.exe
*.o
*.a
node_modules/
dist/
build/

# Exclude docs or test folders
docs/
test/
```

#### Pattern B: Whitelist Mode (Recommended for Large Repos)

```gitignore
# 1. Ignore everything under root
/*

# 2. Re-include target directories and subfiles
!/src/
!/src/**
!/inc/
!/inc/**

# 3. Filter specific files inside whitelisted directories
src/legacy/
*.tmp
```

#### Pattern C: Re-include Files in Excluded Folders

> **Git Rule**: You cannot re-include a file if its parent directory is excluded. Always exclude folder contents instead of the directory itself:

```gitignore
# Correct: Exclude folder contents, then re-include specific file
test/*
!test/main_test.c
```

---

<div id="chinese-documentation"></div>

## 🇨🇳 中文说明

**merge-src** 是一款专为大模型（LLM）对话辅助编程设计的轻量级代码上下文打包工具。它可以根据项目根目录下的 `.mergerule` 规则，将整个项目的源代码智能整合成一个排版清晰、携带语义注释的单一文本文件，方便在网页聊天框中直接提交给 AI。

工具完全遵循业界标准的 `.gitignore` 规则语法，开箱即用，零认知成本。

### ✨ 核心特性

- **🎯 标准 Gitignore 引擎**：由工业级 `go-git` 官方规范驱动，原生支持线性匹配、`!` 否定放行、层级通配符（`*`, `**`）与目录级剪枝，学习成本归零。
- **⚙️ 极致零配置（Zero-Config）**：彻底剔除繁琐的 `key = value` 参数，自动以当前项目目录名命名输出文件（`<项目名>-YYYYMMDD-HHMM.txt`）。
- **🌐 智能编码清洗与转码**：自动检测非 UTF-8 源码（如嵌入式 Keil/ADS 老项目中的 ANSI / GBK / GB2312 / GB18030），自动剥离 BOM 并无损转码为纯正 UTF-8，彻底终结 AI 乱码噩梦。
- **🎨 8 大阵营多语言注释自适应**：全自动识别绝大多数主流、现代新兴及嵌入式底层语言后缀（C/C++、Python、SQL、Lisp/汇编、HTML、LaTeX 等），精准插入对应风格的分割注释线。
- **🛡️ 内置安全防护**：系统自动排除 `.git/` 版本库、`.mergerule`、`.gitignore` 及输出合并文件自身，杜绝自吞噬死循环与元数据污染。

### 📦 安装

#### 方式 1: 使用 Go 直接安装 (推荐)

```bash
go install github.com/vagra/merge-src@latest
```

#### 方式 2: 源码编译

```bash
git clone https://github.com/vagra/merge-src.git
cd merge-src
go build
```

### 🚀 快速上手

1. 进入需要打包的代码项目根目录。
2. 创建规则文件 `.mergerule`（语法与 `.gitignore` 完全一致）。
3. 执行合并命令：

```bash
merge-src
```

### 📝 规则编写指南 (.mergerule)

规则严格遵循标准 Gitignore 规范：**自上而下逐行扫描，后声明的规则覆盖前规则（Last Match Wins）**。

#### 模式 A：黑名单模式（默认全部合并，排除特定杂项）

```gitignore
# 排除编译产物与二进制库
*.exe
*.o
*.a
*.bin

# 排除依赖与构建目录
node_modules/
dist/
build/
bin/

# 排除文档或多媒体资源
docs/
assets/
```

#### 模式 B：白名单模式（大型复杂项目最佳实践）

```gitignore
# 1. 忽略根目录下的所有内容 (IDE 缓存与杂项直接剪枝，扫描极快)
/*

# 2. 显式放行核心源码目录
!/src/
!/src/**
!/inc/
!/inc/**

# 3. 在放行目录中进一步细化排除
src/JPEGlib/
*.tmp
```

#### 模式 C：在已排除目录中“捡回”特定文件

> **核心铁律**：如果父目录整体被排除，其子文件无法通过 `!` 重新包含。必须排除目录下的内容，而非排除目录本身：

```gitignore
# 正确写法：排除 test 目录内容，但单独捡回 main_test.c
test/*
!test/main_test.c
```
