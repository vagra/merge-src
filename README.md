# merge-src: AI Code Context Builder

![License](https://img.shields.io/badge/license-MIT-blue.svg) ![Go](https://img.shields.io/badge/go-1.20+-00ADD8.svg)

> **Navigation**: [🇺🇸 English](#english-documentation) | [🇨🇳 中文说明](#chinese-documentation)

<div id="english-documentation"></div>

## 🇺🇸 English Documentation

**merge-src** is a CLI tool designed to help developers build **Code Context** for LLMs (Large Language Models). It intelligently merges source code from your project into a single text file based on a configuration file (`.mergerule`), automatically adding file path comments and separators.

Unlike `cat` or `find`, its core goal is **Validation** — ensuring the context you feed to AI is complete and accurate.

### ✨ Features

- **🎯 Precise Control**: Define rules via `.mergerule`, supporting recursion, wildcards, and exact paths.
- **🕵️‍♀️ Missing Detection**: The killer feature. If a rule requires a file that doesn't exist, it alerts you with **RED** highlights immediately.
- **🧠 Smart Filtering**: Directory rules automatically respect your `extensions` setting (e.g., only `.c`, `.h`).
- **🎨 Smart Syntax**: Automatically detects file types to apply the correct comment style (e.g., `//` for C, `#` for Python, `<!-- -->` for Quarto). Configured style acts as a fallback.
- **⚡️ Zero Dependency**: Single binary, cross-platform.

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
2. Create a config file `.mergerule` (copy from `.mergerule.example_en`).
3. Run:

```bash
merge-src
```

### 📝 Configuration (.mergerule)

#### Global Settings

```ini
# Auto-filter files by extension (only applies to Directory Mode rules)
extensions = .c, .h, .cpp

# Default/Fallback comment style (used if extension is unknown)
style = c

# Output filename prefix
output_prefix = my-code-analysis
```

#### Path Rules (The Logic)

Rules follow the **Longest Match Wins** principle. A more specific (longer path) rule overrides a general one.

- `+`: Include
- `-`: Exclude

**5 Matching Modes:**

| Syntax | Mode | Recursive? | Check Extensions? | Description |
| :--- | :--- | :--- | :--- | :--- |
| **`+path/`** | **Directory** | ✅ Yes | ✅ Yes | Recursively includes files matching `extensions` in `path`. |
| **`+path/**/pat`** | **Recur. Glob** | ✅ Yes | ❌ No | Recursively includes files matching `pat` in `path`. |
| **`+pat`** | **Global Glob** | ✅ Yes | ❌ No | Recursively includes files matching `pat` in **any directory**. |
| **`+path/pat`** | **Flat Glob** | ❌ No | ❌ No | Includes files matching `pat` in the **current directory only**. |
| **`+path/file`** | **Exact File** | ❌ No | ❌ No | Includes this specific file. **Errors if missing.** |

---

<div id="chinese-documentation"></div>

## 🇨🇳 中文说明

**merge-src** 是一个专为大模型（LLM）辅助编程设计的命令行工具。它可以根据配置文件（`.mergerule`）智能地将项目中的源代码合并为一个单一的文本文件。

它的核心设计目标是**“验证性”** —— 确保你喂给 AI 的代码上下文是完整且准确的，防止“幻觉”上下文。

### ✨ 核心特性

- **🎯 精确控制**：通过 `.mergerule` 配置包含/排除规则，支持递归、通配符和精确文件指定。
- **🕵️‍♀️ 缺失检测**：杀手级功能。如果规则中明确要求的文件未找到，工具会立刻用**红色高亮**报警。
- **🧠 智能过滤**：目录模式下会自动根据 `extensions` 过滤无关文件（如二进制文件）。
- **🎨 智能语法**：自动根据文件后缀（如 .py, .qmd）选择正确的注释风格，生成的分割线不会破坏代码语法。仅在无法识别时使用默认配置。
- **⚡️ 零依赖**：单文件 Go 二进制，跨平台。

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

### 🚀 快速开始

1. 进入你的代码项目目录。
2. 创建配置文件 `.mergerule` (复制 `.mergerule.example_cn` 并重命名)。
3. 运行：

```bash
merge-src
```

### 📝 配置文件规范 (.mergerule)

#### 全局设置

```ini
# 源码后缀 (仅对目录模式生效)
extensions = .c, .h, .cpp

# 默认/保底注释风格 (当后缀无法识别时使用)
style = c

# 输出文件名前缀
output_prefix = my-code-analysis
```

#### 路径规则逻辑

规则遵循 **“最长匹配优先” (Longest Match Wins)** 原则。

- `+` 表示包含 (Include)
- `-` 表示排除 (Exclude)

**5 种匹配模式：**

| 语法格式 | 模式名称 | 是否递归? | 检查全局后缀? | 说明 |
| :--- | :--- | :--- | :--- | :--- |
| **`+path/`** | **目录模式** | ✅ 是 | ✅ 是 | 递归包含 `path` 下所有符合 `extensions` 的文件。 |
| **`+path/**/pat`** | **递归通配** | ✅ 是 | ❌ 否 | 递归包含 `path` 下所有匹配模式 `pat` 的文件。 |
| **`+pat`** | **全局通配** | ✅ 是 | ❌ 否 | 递归包含**任意目录**下匹配模式 `pat` 的文件。 |
| **`+path/pat`** | **平铺通配** | ❌ 否 | ❌ 否 | 仅包含 `path` **当前层级**下匹配 `pat` 的文件。 |
| **`+path/file`** | **精确文件** | ❌ 否 | ❌ 否 | 精确包含此文件。如果文件不存在，会**报错**。 |

