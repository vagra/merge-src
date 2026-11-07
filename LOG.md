# merge-src - 开发履历

## Phase 2: Gitignore Specification Migration & Zero-Config Architecture (2026-11-07) - [Completed] (v2.0.0 & v2.0.1)
**目标**: 废弃认知成本高昂的私有规则语法，全面迁移至标准 Gitignore 规范；建立零配置工作流；彻底根治嵌入式/历史老项目中的编码兼容与中文乱码问题。
*   **[Engine] 全面拥抱标准 Gitignore 规范 (`internal/config.go`, `internal/engine.go`)**:
    *   **范式转变**: 废除原基于字符串长度优先（Longest Match Wins）的私有规则，迁移至工业级标准 `github.com/go-git/go-git/v5/plumbing/format/gitignore` 规范。
    *   **语法统一**: 原生支持线性自上而下匹配、后规则覆盖前规则（Last Match Wins）、支持 `!` 模式放行/捡回以及标准通配符（`*`, `**`）。
    *   **目录级剪枝 (Pruning)**: 遇到被排除的目录直接截断深入扫描，大幅优化大型复杂项目的遍历合并性能。
    *   **内置安全防护**: 自动排除 `.git/`、`.mergerule`、`.gitignore` 及输出目标文件自身，杜绝自吞噬死循环与版本库元数据污染。
*   **[Workflow] 极致简洁的零配置架构 (`main.go`, `internal/config.go`)**:
    *   **消除冗余配置**: 彻底废除 `output_prefix`、`style`、`extensions` 等所有 key-value 配置项，使 `.mergerule` 呈现为一个纯粹的 Gitignore 规则文件。
    *   **动态环境自适应**: 自动提取当前项目根目录名作为输出文件名前缀；多语言注释探测扩充覆盖至 8 大阵营主流、嵌入式着色器及现代新兴语言。
*   **[Encoding] 混合编码清洗与强制 UTF-8 归一化 (`internal/engine.go`)**:
    *   **病灶**: 针对 Windows 嵌入式老项目（如 Keil/ADS/IAR）中的 ANSI/GBK 源码，直接按字节写入会导致生成“UTF-8 + GBK 混合编码”怪胎，造成 AI 网页与编辑器中文乱码。
    *   **解决**: 引入 `golang.org/x/text`，在正文写入前进行全自动编码检测。自动剥离 UTF-8 BOM 头，并利用 GB18030（向下兼容 GBK/GB2312）解码器动态无损转码为纯正合法的 UTF-8 文本流。

## Phase 1: Global Rule Intuition & Engine Refactor (2026-03-20) - [Completed]
**目标**: 解决路径匹配规则过于机械化、反直觉的问题，引入符合 `.gitignore` 行业惯例的“全局递归”匹配逻辑，提升配置文件的编写效率与可维护性。
*   **[Core] 规则解析器逻辑重构 (`internal/config.go`)**:
    *   **病灶**: 原解析逻辑将不带斜杠的规则（如 `-file.h`）等同于根目录下的精确文件匹配。若要排除全库的同名文件，被迫使用 `-./**/file.h` 这种繁琐且高认知的表达方式，严重违背了“简洁至上”的原则。
    *   **解决**: 引入**无斜杠探测机制**。在解析阶段，若发现 `Pattern` 中不含 `/` 字符，则自动将其判定为“全局递归模式”：
        1. 强制设置 `Recursive = true`。
        2. 将 `BaseDir` 设为空字符串（即代表全项目路径）。
    *   **兼容性设计**: 保留了通过 `+./file.c`（显式带斜杠）来实现“仅匹配当前层级”的平铺匹配能力，实现了“直觉化全库匹配”与“精确单点匹配”的完美解耦。
*   **[Engine] 验证逻辑与缺失报警优化 (`internal/engine.go`)**:
    *   **逻辑修正**: 在 `initRules` 初始化阶段，对“文件缺失报警”的触发条件进行了原子级修正。
    *   **实现**: 明确区分“意图包含”与“存在性断言”。
        1. 只有当规则被定义为 **非递归 (Flat)** 且 **非通配 (No Glob)** 的精确路径时（如 `+src/main.c`），才将其加入 `ExpectationsByDir` 期待列表。
        2. 对于新增的全局递归规则（如 `+README.md`），由于其物理分布的不确定性，不再强制要求在根目录进行存在性校验，从而斩断了由于逻辑泛化导致的伪“红字缺失”报错。
*   **[Docs] 匹配模式五位一体化 (`README.md`)**:
    *   **改进**: 在中英文文档中同步更新了路径匹配矩阵，正式确立了 **Global Glob (全局通配)** 作为第 5 种标准匹配模式的地位。
    *   **价值**: 明确了“最长路径匹配优先”与“无斜杠即全局”的底层逻辑，为开发者构建 AI 代码上下文提供了更清晰的决策指引。
