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

var anchorIDs = []string{
	"q8fkHP9olS", // COMUNIDADE
	"MU2bRRm5e2", // Comercial
	"GOYM-qu2xH", // PROSPECTING SYSTEM
	"mf29oHB4mw", // INFRAESTRUTURA
	"RUUUdJ9yoR", // Onboarding Showrunner
	"b-baf95a0d86", // LightRAG
	"b-59914bfaac", // RAGFlow
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

type anchorReport struct {
	ID     string   `json:"id"`
	Text   string   `json:"text,omitempty"`
	Parent string   `json:"parent,omitempty"`
	XYWH   string   `json:"xywh,omitempty"`
	Chain  []string `json:"chain,omitempty"`
}

type report struct {
	Timestamp string                  `json:"timestamp"`
	Blocks    int                     `json:"blocks"`
	Notes     int                     `json:"notes"`
	Bounds    map[string]float64       `json:"bounds,omitempty"`
	Anchors   map[string]anchorReport  `json:"anchors,omitempty"`
	Error     string                  `json:"error,omitempty"`
}

func main() {
	cfg, err := config.Load("")
	must(err)
	engine, err := yjs.NewEngine()
	must(err)

	timestamps := readTimestamps(`D:\tmp\affine-history-200.json`)
	if len(os.Args) > 1 {
		timestamps = os.Args[1:]
	}

	var reports []report
	for _, ts := range timestamps {
		reports = append(reports, auditSnapshot(cfg, engine, ts))
	}
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
	parents := map[string]string{}
	for id, block := range blocks {
		for _, child := range stringSlice(block["sys:children"]) {
			parents[child] = id
		}
	}

	r := report{
		Timestamp: ts,
		Blocks:    len(blocks),
		Anchors:   map[string]anchorReport{},
	}

	var minX, minY, maxX, maxY float64
	hasBounds := false
	for _, block := range blocks {
		if str(block["sys:flavour"]) != "affine:note" {
			continue
		}
		r.Notes++
		xywh := str(block["prop:xywh"])
		x, y, w, h, ok := parseXYWH(xywh)
		if !ok {
			continue
		}
		if !hasBounds {
			minX, minY, maxX, maxY = x, y, x+w, y+h
			hasBounds = true
			continue
		}
		minX = min(minX, x)
		minY = min(minY, y)
		maxX = max(maxX, x+w)
		maxY = max(maxY, y+h)
	}
	if hasBounds {
		r.Bounds = map[string]float64{
			"minX": minX, "minY": minY, "maxX": maxX, "maxY": maxY,
			"width": maxX - minX, "height": maxY - minY,
		}
	}

	for _, id := range anchorIDs {
		r.Anchors[id] = anchor(id, blocks, parents)
	}
	return r
}

func anchor(id string, blocks map[string]map[string]any, parents map[string]string) anchorReport {
	block, ok := blocks[id]
	if !ok {
		return anchorReport{ID: id}
	}
	a := anchorReport{ID: id, Text: str(block["prop:text"])}
	seen := map[string]bool{}
	for cur := id; cur != ""; cur = parents[cur] {
		if seen[cur] {
			break
		}
		seen[cur] = true
		a.Chain = append(a.Chain, cur)
		p := parents[cur]
		if p == "" {
			break
		}
		parent := blocks[p]
		if str(parent["sys:flavour"]) == "affine:note" {
			a.Parent = p
			a.XYWH = str(parent["prop:xywh"])
			break
		}
	}
	return a
}

func stringSlice(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s := str(item); s != "" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func parseXYWH(s string) (float64, float64, float64, float64, bool) {
	var vals []float64
	if err := json.Unmarshal([]byte(s), &vals); err != nil || len(vals) != 4 {
		return 0, 0, 0, 0, false
	}
	return vals[0], vals[1], vals[2], vals[3], true
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
