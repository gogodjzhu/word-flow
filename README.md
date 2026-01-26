# Word-Flow

[中文文档](README_zh.md) | **English**

Word-Flow is a powerful terminal-based dictionary and vocabulary learning application built for developers. It combines multiple dictionary sources, AI-powered translations, and a notebook system to help you master new words directly from your command line.

> **Note**: This project is inspired by [Wudao-dict](https://github.com/ChestnutHeng/Wudao-dict) with enhanced dictionary lookup and translation capabilities. It is developed purely in Go with no external dependencies—ready to use out of the box.

## Features

- **Multi-Source Dictionary Lookup**:
  - **Youdao**: Concise online definitions and translations (Default).
  - **Etymonline**: Explore word origins and history.
  - **ECDICT**: Massive offline dictionary support.
  - **Merriam-Webster**: Authoritative definitions (requires API key).
  - **LLM**: AI-powered definitions and explanations.

- **AI Translation**:
  - Translate text to Chinese using Large Language Models (LLM).
  - Supports streaming output.
  - Reference mode (`--ref`) to show original text alongside translation.

- **Vocabulary Notebook**:
  - Save words to your local notebook.
  - Review words (Future support: spaced repetition and exam mode).

- **Cross-Platform**: Works on macOS, Linux, and Windows.

## Installation

### Download Binary (Recommended)

Download the latest release for your platform from the [Releases](https://github.com/gogodjzhu/word-flow/releases) page.

1. Download the archive (e.g., `wordflow-linux-amd64.tar.gz`).
2. Extract the binary.
3. Move `wordflow` to a directory in your `PATH` (e.g., `/usr/local/bin`).

```bash
tar -xzf wordflow-linux-amd64.tar.gz
sudo mv wordflow /usr/local/bin/
```

### Build from Source

Requirements: Go 1.22+

```bash
# Clone the repository
git clone https://github.com/gogodjzhu/word-flow.git
cd word-flow

# Build
go build -o wordflow cmd/wordflow/main.go

# Install (optional, move to your PATH)
mv wordflow /usr/local/bin/
```

## Usage

### Dictionary Lookup (`dict`)

Look up a word using the default dictionary (Youdao):
```bash
wordflow dict "ephemeral"
```

Specify a dictionary source:
```bash
wordflow dict "legacy" -d etymonline
wordflow dict "complex" -d llm
```

List available dictionaries:
```bash
wordflow dict -l
```

### AI Translation (`trans`)

Translate a sentence:
```bash
wordflow trans "Hello world, this is a test."
```

Use pipe input with reference display:
```bash
echo "Software engineering is the application of engineering to the development of software." | wordflow trans --stdin --ref
```

### Vocabulary Notebook (`notebook`)

Words looked up via the `dict` command are automatically saved to your notebook.

Review words in your notebook:
```bash
wordflow notebook
```

## Configuration

Word-Flow uses a YAML configuration file located at `~/.config/wordflow/config.yaml`. The file is automatically created on the first run.

### Supported Dictionaries

| Source | Type | Description |
|--------|------|-------------|
| `youdao` | Online | Free, concise definitions. |
| `etymonline` | Online | Free, etymology and word origins. |
| `ecdict` | Offline | Free, requires downloading `stardict.db`. |
| `mwebster` | API | Requires Merriam-Webster API key. |
| `llm` | API | AI definitions (OpenAI compatible). |

### Example Configuration

To use LLM features or API-based dictionaries, edit your config file:

```yaml
dict:
  default: youdao
  parameters:
    # LLM Settings
    llm.api_key: "your-api-key"
    llm.url: "https://api.openai.com/v1"
    llm.model: "gpt-3.5-turbo"
    
    # Merriam-Webster
    mwebster.key: "your-dictionary-api-key"
    
    # ECDICT (Offline)
    ecdict.dbfilename: "/path/to/stardict.db"
```

## License

Apache License 2.0
