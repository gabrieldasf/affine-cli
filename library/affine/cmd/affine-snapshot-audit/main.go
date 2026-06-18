package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"affine-pp-cli/internal/config"
	"affine-pp-cli/internal/yjs"
)

const (
	workspaceID = "727cc066-a25e-4560-b68d-414b67cbc5c8"
	docID       = "B6pvUw-r5SSfWKam-wncU"
)

var terms = []string{
	"app", "apps", "comercial", "comunidade", "conteúdo", "conteudo",
	"currículo", "curriculo", "infra", "infraestrutura", "operacional",
	"financeiro", "prospec", "carta", "clube", "coolify", "hetzner",
	"contabo", "traefik", "lightrag", "ragflow", "papermark", "documenso",
}

type historyResponse struct {
	Data struct {
		Data struct {
			Workspace struct {
				Histories []struct {
					Timestamp string `json:"timestamp"`
				} `json:"histories"`
			} `json:"workspace"`
		} `json:"data"`
	} `json:"data"`
}

type report struct {
	Timestamp   string         `json:"timestamp"`
	Blocks      int            `json:"blocks"`
	Notes       int            `json:"notes"`
	Images      int            `json:"images"`
	Frames      int            `json:"frames"`
	TextBlocks  int            `json:"text_blocks"`
	TextChars   int            `json:"text_chars"`
	Hits        map[string]int `json:"hits"`
	SampleTexts []string       `json:"sample_texts,omitempty"`
	Error       string         `json:"error,omitempty"`
}

func main() {
	cfg, err := config.Load("")
	must(err)
	engine, err := yjs.NewEngine()
	must(err)

	timestamps := readTimestamps(`D:\tmp\affine-history-200.json`)
	var reports []report
	for _, ts := range timestamps {
		reports = append(reports, auditSnapshot(cfg, engine, ts))
	}
	sort.SliceStable(reports, func(i, j int) bool {
		if reports[i].Blocks == reports[j].Blocks {
			return reports[i].TextChars > reports[j].TextChars
		}
		return reports[i].Blocks > reports[j].Blocks
	})

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	must(enc.Encode(reports))
}

func readTimestamps(path string) []string {
	raw, err := os.ReadFile(path)
	must(err)
	raw = []byte(strings.TrimPrefix(string(raw), "\ufeff"))
	var parsed historyResponse
	must(json.Unmarshal(raw, &parsed))
	seen := map[string]bool{}
	var out []string
	for _, h := range parsed.Data.Data.Workspace.Histories {
		if h.Timestamp != "" && !seen[h.Timestamp] {
			seen[h.Timestamp] = true
			out = append(out, h.Timestamp)
		}
	}
	return out
}

func auditSnapshot(cfg *config.Config, engine *yjs.Engine, ts string) report {
	base := strings.TrimSuffix(strings.TrimSuffix(cfg.BaseURL, "/graphql"), "/")
	url := fmt.Sprintf("%s/api/workspaces/%s/docs/%s/histories/%s", base, workspaceID, docID, ts)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return report{Timestamp: ts, Error: err.Error()}
	}
	if cfg.AuthHeader() != "" {
		req.Header.Set("Authorization", cfg.AuthHeader())
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return report{Timestamp: ts, Error: err.Error()}
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return report{Timestamp: ts, Error: fmt.Sprintf("%s %s", res.Status, string(body))}
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return report{Timestamp: ts, Error: err.Error()}
	}
	doc, err := engine.ApplyBase64Update(base64.StdEncoding.EncodeToString(raw))
	if err != nil {
		return report{Timestamp: ts, Error: err.Error()}
	}
	blocks, err := engine.ReadBlocks(doc)
	if err != nil {
		return report{Timestamp: ts, Error: err.Error()}
	}
	return summarize(ts, blocks)
}

func summarize(ts string, blocks map[string]map[string]any) report {
	r := report{Timestamp: ts, Blocks: len(blocks), Hits: map[string]int{}}
	var samples []string
	for _, block := range blocks {
		flavour := lower(block["sys:flavour"])
		switch {
		case strings.Contains(flavour, "note"):
			r.Notes++
		case strings.Contains(flavour, "image"):
			r.Images++
		case strings.Contains(flavour, "frame"):
			r.Frames++
		}
		text := str(block["prop:text"])
		if text == "" {
			continue
		}
		r.TextBlocks++
		r.TextChars += len([]rune(text))
		textLower := strings.ToLower(text)
		matched := false
		for _, term := range terms {
			if strings.Contains(textLower, term) {
				r.Hits[term]++
				matched = true
			}
		}
		if matched && len(samples) < 24 {
			samples = append(samples, clip(text, 180))
		}
	}
	sort.Strings(samples)
	r.SampleTexts = samples
	return r
}

func lower(v any) string { return strings.ToLower(str(v)) }

func str(v any) string {
	s, _ := v.(string)
	return s
}

func clip(s string, n int) string {
	rs := []rune(strings.Join(strings.Fields(s), " "))
	if len(rs) <= n {
		return string(rs)
	}
	return string(rs[:n]) + "..."
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
