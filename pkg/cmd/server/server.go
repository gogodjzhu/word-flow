package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gogodjzhu/word-flow/pkg/cmdutil"
	"github.com/gogodjzhu/word-flow/pkg/dict"
	"github.com/spf13/cobra"
)

var htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Word}} - Word Flow</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px 12px;
        }
        .container {
            max-width: 700px;
            margin: 0 auto;
        }
        .search-box {
            background: white;
            border-radius: 12px;
            padding: 16px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            margin-bottom: 20px;
        }
        .search-box form {
            display: flex;
            gap: 8px;
            flex-direction: row;
        }
        .search-box input {
            flex: 1;
            min-width: 0;
            padding: 12px 14px;
            font-size: 16px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            outline: none;
            transition: border-color 0.2s;
        }
        .search-box input:focus {
            border-color: #667eea;
        }
        .search-box button {
            padding: 12px 20px;
            font-size: 16px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            transition: background 0.2s;
            white-space: nowrap;
        }
        .search-box button:hover {
            background: #5568d3;
        }
        .result {
            background: white;
            border-radius: 12px;
            padding: 24px 20px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .word {
            font-size: 28px;
            font-weight: 700;
            color: #333;
            margin-bottom: 6px;
        }
        .source {
            font-size: 13px;
            color: #888;
            margin-bottom: 14px;
        }
        .phonetics {
            margin-bottom: 18px;
            display: flex;
            flex-wrap: wrap;
            gap: 8px;
        }
        .phonetic {
            display: inline-flex;
            align-items: center;
            background: #f5f5f5;
            padding: 8px 12px;
            border-radius: 6px;
            font-size: 14px;
            color: #555;
            gap: 6px;
        }
        .phonetic button {
            background: transparent;
            color: #667eea;
            border: none;
            padding: 2px 6px;
            font-size: 12px;
            cursor: pointer;
            font-family: inherit;
            font-weight: 600;
        }
        .phonetic button:hover {
            color: #5568d3;
            text-decoration: underline;
        }
        .meanings {
            margin-bottom: 18px;
        }
        .meaning {
            margin-bottom: 14px;
        }
        .pos {
            display: inline-block;
            background: #e8f5e9;
            color: #2e7d32;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 13px;
            font-weight: 600;
            margin-right: 6px;
        }
        .definitions {
            margin-top: 6px;
            line-height: 1.6;
            color: #444;
            font-size: 15px;
        }
        .example {
            margin-top: 8px;
            padding: 10px;
            background: #f8f9fa;
            border-left: 3px solid #667eea;
            font-style: italic;
            color: #666;
            font-size: 14px;
            line-height: 1.5;
        }
        .examples {
            margin-top: 14px;
        }
        .examples .example {
            margin-bottom: 8px;
        }
        .error {
            background: #ffebee;
            color: #c62828;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }

        @media (max-width: 480px) {
            body {
                padding: 12px 8px;
            }
            .search-box {
                padding: 12px;
            }
            .search-box form {
                flex-direction: column;
            }
            .search-box button {
                width: 100%;
                padding: 14px;
            }
            .result {
                padding: 16px 14px;
            }
            .word {
                font-size: 24px;
            }
            .phonetics {
                gap: 6px;
            }
            .phonetic {
                font-size: 12px;
                padding: 6px 8px;
            }
            .phonetic button {
                padding: 4px 8px;
                font-size: 11px;
            }
            .pos {
                font-size: 12px;
            }
            .definitions {
                font-size: 14px;
            }
            .example {
                font-size: 13px;
                padding: 8px;
            }
        }

        @media (min-width: 481px) and (max-width: 768px) {
            body {
                padding: 30px 16px;
            }
            .search-box {
                padding: 14px;
            }
            .result {
                padding: 20px 18px;
            }
            .word {
                font-size: 26px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="search-box">
            <form method="GET" action="/dict">
                <input type="text" name="word" value="{{.QueryWord}}" placeholder="Enter a word..." autofocus>
                <button type="submit">Search</button>
            </form>
        </div>
        {{if .Error}}
        <div class="result">
            <div class="error">{{.Error}}</div>
        </div>
        {{else if .Word}}
        <div class="result">
            <div class="word">{{.Word}}</div>
            <div class="source">{{.Source}}</div>
            
            {{if .Phonetics}}
            <div class="phonetics">
                {{range .Phonetics}}
                <span class="phonetic">
                    <button onclick="new Audio('{{.Audio}}').play()">[{{.LanguageCode}}] {{.Text}}</button>
                </span>
                {{end}}
            </div>
            {{end}}
            
            {{range .Meanings}}
            <div class="meaning">
                {{if .PartOfSpeech}}<span class="pos">{{.PartOfSpeech}}</span>{{end}}
                <div class="definitions">{{.Definitions}}</div>
                {{if .Examples}}
                <div class="examples">
                    {{range .Examples}}
                    <div class="example">eg. {{.}}</div>
                    {{end}}
                </div>
                {{end}}
            </div>
            {{end}}
            
            {{if .Examples}}
            <div class="examples">
                {{range .Examples}}
                <div class="example">{{.}}</div>
                {{end}}
            </div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>
`

type TemplateData struct {
	QueryWord string
	Word      string
	Source    string
	Phonetics []dictPhonetic
	Meanings  []dictMeaning
	Examples  []string
	Error     string
}

type dictPhonetic struct {
	LanguageCode string
	Text         string
	Audio        string
}

type dictMeaning struct {
	PartOfSpeech string
	Definitions  string
	Examples     []string
}

func NewCmdServer(f *cmdutil.Factory) (*cobra.Command, error) {
	var port int
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start HTTP server for word lookup",
		Long:  "Start an HTTP server that provides a web interface for dictionary lookups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServer(f, port)
		},
	}
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for HTTP server")
	return cmd, nil
}

func startServer(f *cmdutil.Factory, port int) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	dictionary, err := dict.NewDict(cfg.Dict)
	if err != nil {
		return err
	}

	tmpl := template.Must(template.New("dict").Parse(htmlTemplate))

	http.HandleFunc("/dict", func(w http.ResponseWriter, r *http.Request) {
		word := r.URL.Query().Get("word")
		if word == "" {
			tmpl.Execute(w, TemplateData{QueryWord: word})
			return
		}

		word = strings.TrimSpace(word)
		wordItem, err := dictionary.Search(word)
		if err != nil {
			tmpl.Execute(w, TemplateData{QueryWord: word, Error: err.Error()})
			return
		}

		notebookName := r.URL.Query().Get("notebook")
		if notebookName == "" {
			notebookName = cfg.Notebook.Default
		}

		originalNotebook := cfg.Notebook.Default
		cfg.Notebook.Default = notebookName

		notebookConfig, err := cfg.Notebook.GetConfig()
		if err != nil {
			tmpl.Execute(w, TemplateData{QueryWord: word, Error: err.Error()})
			return
		}
		notebook, err := dict.OpenNotebook(notebookConfig)
		if err != nil {
			tmpl.Execute(w, TemplateData{QueryWord: word, Error: err.Error()})
			return
		}
		cfg.Notebook.Default = originalNotebook

		if _, err := notebook.Mark(wordItem.Word, dict.Learning, wordItem); err != nil {
			tmpl.Execute(w, TemplateData{QueryWord: word, Error: err.Error()})
			return
		}

		phonetics := make([]dictPhonetic, len(wordItem.WordPhonetics))
		for i, p := range wordItem.WordPhonetics {
			phonetics[i] = dictPhonetic{
				LanguageCode: p.LanguageCode,
				Text:         p.Text,
				Audio:        p.Audio,
			}
		}

		meanings := make([]dictMeaning, len(wordItem.WordMeanings))
		for i, m := range wordItem.WordMeanings {
			meanings[i] = dictMeaning{
				PartOfSpeech: m.PartOfSpeech,
				Definitions:  m.Definitions,
				Examples:     m.Examples,
			}
		}

		data := TemplateData{
			QueryWord: word,
			Word:      wordItem.Word,
			Source:    wordItem.Source,
			Phonetics: phonetics,
			Meanings:  meanings,
			Examples:  wordItem.Examples,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/dict" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		tmpl.Execute(w, TemplateData{})
	})

	fmt.Printf("Starting word-flow server on http://localhost:%d\n", port)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  GET /              - Web interface\n")
	fmt.Printf("  GET /dict?word=<word>&notebook=<notebook> - Lookup word (supports notebook param)\n")
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
