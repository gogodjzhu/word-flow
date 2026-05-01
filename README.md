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
  - Review and manage words in a TUI (open cached translations, delete items).
  - Spaced repetition review with **FSRS** scheduling.

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

### Install via npm

```bash
npm install -g @gogodjzhu/wordflow@latest
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

#### `notebook review`

Use this to review saved words in a clean TUI list, check cached translations, and remove items you no longer need.

Review words in your notebook (TUI list):
```bash
wordflow notebook review
```

#### `notebook exam`

Use this for a structured review session that focuses only on words due for practice.

Start a spaced-repetition exam session for due words:
```bash
wordflow notebook exam
```

#### `notebook import`

Use this to import a TSV word list into the notebook with lookup during import.

Import words into your notebook from a TSV file:
```bash
wordflow notebook import -i words.tsv
```

## Configuration

Word-Flow uses a YAML configuration file located at `~/.config/wordflow/config.yaml` (or `$WORDFLOW_HOME/config.yaml`). The file is automatically created on the first run with commented defaults.

### Config Command

Manage your configuration from the command line:

```bash
# View current configuration (with env var overrides applied)
wordflow config view

# Get a single value
wordflow config get dict.llm.api_key

# Set a value
wordflow config set dict.llm.api_key sk-xxx
wordflow config set dict.default llm

# Show config file path
wordflow config path

# Re-generate config file with defaults (use --force to overwrite)
wordflow config init
wordflow config init --force

# Use a custom config file
wordflow --config /path/to/config.yaml dict hello
```

### Environment Variables

All config values can be overridden via environment variables with the `WORDFLOW_` prefix:

```bash
WORDFLOW_DICT_LLM_API_KEY=sk-xxx wordflow trans "hello"
WORDFLOW_DICT_DEFAULT=llm wordflow dict "ephemeral"
```

### Supported Dictionaries

| Source | Type | Description |
|--------|------|-------------|
| `youdao` | Online | Free, concise definitions. |
| `etymonline` | Online | Free, etymology and word origins. |
| `ecdict` | Offline | Free, requires downloading `stardict.db`. |
| `mwebster` | API | Requires Merriam-Webster API key. |
| `llm` | API | AI definitions (OpenAI compatible). |

### Example Configuration

```yaml
version: v1

dict:
  default: youdao

  youdao: {}

  llm:
    api_key: ""           # Required. Set via WORDFLOW_DICT_LLM_API_KEY or wordflow config set
    url: ""               # Required. Full API endpoint URL, not base URL
    model: ""             # Required. LLM model name
    timeout: 30s
    max_tokens: 2000
    temperature: 0.3

  ecdict:
    # db_filename: ""    # Defaults to <WORDFLOW_HOME>/stardict.db if empty

  etymonline: {}

  mwebster:
    # key: ""            # Required if using mwebster

notebook:
  default: default

  settings:
    # basepath: ""       # Defaults to <WORDFLOW_HOME>/notebooks if empty
    max_reviews_per_session: 50
    new_cards_per_day: 20
```

## License

Apache License 2.0
