# Word-Flow

**中文文档** | [English](README.md)

Word-Flow 是一个专为开发者打造的强大终端字典与词汇学习工具。它集成了多种字典源、AI 翻译功能以及单词本系统，帮助你在命令行中高效查询和掌握新词汇。

> **说明**: 本项目受到 [Wudao-dict](https://github.com/ChestnutHeng/Wudao-dict) 项目的启发，在此基础上对查词和翻译能力做了进一步增强。本项目完全使用 Go 语言开发，无任何第三方依赖，开箱即用。

## 功能特性

- **多源字典查询**:
  - **有道 (Youdao)**: 在线词典，提供简明释义与翻译（默认）。
  - **词源字典 (Etymonline)**: 探索单词的起源与历史。
  - **ECDICT**: 支持超大离线词库查询。
  - **韦氏词典 (Merriam-Webster)**: 权威英语释义（需 API Key）。
  - **LLM**: 利用大语言模型提供智能释义与详解。

- **AI 智能翻译**:
  - 使用大语言模型 (LLM) 将英文文本翻译为中文。
  - 支持流式输出，实时显示结果。
  - 对照模式 (`--ref`)：同时显示原文与译文，方便双语阅读。

- **单词本与记忆**:
  - 将生词保存到本地单词本。
  - TUI 列表复习与管理（查看已缓存翻译、删除条目）。
  - 基于 **FSRS** 间隔重复算法的复习计划。

- **跨平台支持**: 完美运行于 macOS, Linux 和 Windows。

## 安装指南

### 下载可执行文件 (推荐)

请访问 [Releases](https://github.com/gogodjzhu/word-flow/releases) 页面下载对应您操作系统的最新版本。

1. 下载压缩包 (例如 `wordflow-linux-amd64.tar.gz`)。
2. 解压获取可执行文件。
3. 将 `wordflow` 移动到系统 `PATH` 路径下 (例如 `/usr/local/bin`)。

```bash
tar -xzf wordflow-linux-amd64.tar.gz
sudo mv wordflow /usr/local/bin/
```

### 源码编译安装

环境要求: Go 1.22+

```bash
# 克隆仓库
git clone https://github.com/gogodjzhu/word-flow.git
cd word-flow

# 编译
go build -o wordflow cmd/wordflow/main.go

# 安装 (可选：移动到系统 PATH 路径下)
mv wordflow /usr/local/bin/
```

## 使用说明

### 查词 (`dict`)

使用默认字典（有道）查询单词：
```bash
wordflow dict "ephemeral"
```

指定字典源进行查询：
```bash
wordflow dict "legacy" -d etymonline  # 查看词源
wordflow dict "complex" -d llm        # AI 详解
```

列出所有可用字典：
```bash
wordflow dict -l
```

### AI 翻译 (`trans`)

翻译句子：
```bash
wordflow trans "Hello world, this is a test."
```

使用管道输入并开启对照模式：
```bash
echo "Software engineering is the application of engineering to the development of software." | wordflow trans --stdin --ref
```

### 单词本 (`notebook`)

使用 `dict` 命令查询的单词会自动保存到您的单词本中。

复习单词（TUI 列表）：
```bash
wordflow notebook
```

开始一次间隔重复测验（仅抽取到期单词）：
```bash
wordflow notebook -o exam
```

## 配置说明

Word-Flow 使用 YAML 格式的配置文件，默认位于 `~/.config/wordflow/config.yaml`。首次运行程序时会自动生成该文件。

### 支持的字典源

| 字典源 | 类型 | 说明 |
|--------|------|-------------|
| `youdao` | 在线 | 免费，简明释义。 |
| `etymonline` | 在线 | 免费，词源查询。 |
| `ecdict` | 离线 | 免费，需下载 `stardict.db` 数据库文件。 |
| `mwebster` | API | 需要 Merriam-Webster API Key。 |
| `llm` | API | AI 智能释义 (兼容 OpenAI 接口)。 |

### 配置示例

如需使用 LLM 功能或 API 类字典，请编辑配置文件：

```yaml
dict:
  default: youdao
  parameters:
    # LLM 设置
    llm.api_key: "your-api-key"
    llm.url: "https://api.openai.com/v1"
    llm.model: "gpt-3.5-turbo"
    
    # 韦氏词典设置
    mwebster.key: "your-dictionary-api-key"
    
    # ECDICT (离线词库)
    ecdict.dbfilename: "/path/to/stardict.db"

notebook:
  default: default
  parameters:
    # 单词本存储路径
    notebook.basepath: "~/.config/wordflow/notebooks"

    # FSRS（间隔重复）
    fsrs.max_reviews_per_session: 50
    fsrs.new_cards_per_day: 20
```

## 许可证

Apache License 2.0
