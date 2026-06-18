package canvaswrite

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"affine-pp-cli/internal/config"
	"affine-pp-cli/internal/socketio"
	"affine-pp-cli/internal/yjs"
)

type RebuildQuartzoOptions struct {
	WorkspaceID string
	SourceDocID string
	NewDocID    string
	Title       string
	DryRun      bool
	LocalOnly   bool
	SnapshotDir string
}

type RebuildQuartzoResult struct {
	WorkspaceID   string         `json:"workspace_id"`
	SourceDocID   string         `json:"source_doc_id"`
	NewDocID      string         `json:"new_doc_id"`
	Title         string         `json:"title"`
	DryRun        bool           `json:"dry_run"`
	Notes         int            `json:"notes"`
	Images        int            `json:"images"`
	Blocks        int            `json:"blocks"`
	RootPageCount int            `json:"root_page_count"`
	Groups        map[string]int `json:"groups"`
	URL           string         `json:"url"`
}

func RebuildQuartzo(cfg *config.Config, opts RebuildQuartzoOptions) (RebuildQuartzoResult, error) {
	if opts.WorkspaceID == "" {
		opts.WorkspaceID = defaultWorkspaceID
	}
	if opts.SourceDocID == "" {
		opts.SourceDocID = "B6pvUw-r5SSfWKam-wncU"
	}
	if opts.Title == "" {
		opts.Title = "Quartzo Ecosystem v2 - Rebuild"
	}
	if opts.NewDocID == "" {
		opts.NewDocID = "qtz-ecosystem-v2-" + strconv.FormatInt(time.Now().Unix(), 36)
	}

	client, err := connect(cfg, opts.WorkspaceID)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	defer client.Close()

	engine, err := yjs.NewEngine()
	if err != nil {
		return RebuildQuartzoResult{}, err
	}

	sourceLoaded, err := client.LoadDoc(opts.WorkspaceID, opts.SourceDocID)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	if sourceLoaded.Missing == "" {
		return RebuildQuartzoResult{}, fmt.Errorf("source document returned empty snapshot")
	}
	sourceDoc, err := engine.ApplyBase64Update(sourceLoaded.Missing)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	if err := ensureEngineDocIntegrity(engine, sourceDoc, opts.SourceDocID, "before quartzo rebuild"); err != nil {
		return RebuildQuartzoResult{}, err
	}
	rawSummary, err := engine.RunScript(rebuildDocScript(sourceDoc, opts.Title))
	if err != nil {
		return RebuildQuartzoResult{}, fmt.Errorf("rebuild source doc: %w", err)
	}
	var summary struct {
		Blocks int            `json:"blocks"`
		Notes  int            `json:"notes"`
		Images int            `json:"images"`
		Groups map[string]int `json:"groups"`
	}
	if err := json.Unmarshal([]byte(rawSummary), &summary); err != nil {
		return RebuildQuartzoResult{}, fmt.Errorf("parse rebuild summary: %w", err)
	}
	if err := ensureEngineDocIntegrity(engine, sourceDoc, opts.NewDocID, "after local quartzo rebuild"); err != nil {
		return RebuildQuartzoResult{}, err
	}

	result := RebuildQuartzoResult{
		WorkspaceID: opts.WorkspaceID,
		SourceDocID: opts.SourceDocID,
		NewDocID:    opts.NewDocID,
		Title:       opts.Title,
		DryRun:      opts.DryRun,
		Notes:       summary.Notes,
		Images:      summary.Images,
		Blocks:      summary.Blocks,
		Groups:      summary.Groups,
		URL:         fmt.Sprintf("%s/workspace/%s/%s", strings.TrimRight(cfg.BaseURL, "/"), opts.WorkspaceID, opts.NewDocID),
	}
	if opts.DryRun {
		return result, nil
	}

	newDocUpdate, err := engine.EncodeStateAsUpdate(sourceDoc)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}

	rootLoaded, err := client.LoadDoc(opts.WorkspaceID, opts.WorkspaceID)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	if rootLoaded.Missing == "" {
		return RebuildQuartzoResult{}, fmt.Errorf("root document returned empty snapshot")
	}
	rootDoc, err := engine.ApplyBase64Update(rootLoaded.Missing)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	rootStateVector, err := engine.SaveStateVector(rootDoc)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	rawRoot, err := engine.RunScript(registerRootDocScript(rootDoc, opts.NewDocID, opts.Title))
	if err != nil {
		return RebuildQuartzoResult{}, fmt.Errorf("register root doc: %w", err)
	}
	var rootSummary struct {
		PageCount int `json:"pageCount"`
	}
	if err := json.Unmarshal([]byte(rawRoot), &rootSummary); err != nil {
		return RebuildQuartzoResult{}, fmt.Errorf("parse root summary: %w", err)
	}
	rootFullUpdate, err := engine.EncodeStateAsUpdate(rootDoc)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	if opts.SnapshotDir != "" {
		if err := writeRebuildSnapshots(opts.SnapshotDir, opts.WorkspaceID, opts.NewDocID, opts.Title, newDocUpdate, rootFullUpdate); err != nil {
			return RebuildQuartzoResult{}, err
		}
	}
	result.RootPageCount = rootSummary.PageCount
	if opts.LocalOnly {
		return result, nil
	}

	client.PushDocUpdateNoAck(opts.WorkspaceID, opts.NewDocID, newDocUpdate)
	if err := waitForDoc(client, opts.WorkspaceID, opts.NewDocID); err != nil {
		return RebuildQuartzoResult{}, fmt.Errorf("verify rebuilt doc write: %w", err)
	}
	verifiedLoaded, err := client.LoadDoc(opts.WorkspaceID, opts.NewDocID)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	verifiedDoc, err := engine.ApplyBase64Update(verifiedLoaded.Missing)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	if err := ensureEngineDocIntegrity(engine, verifiedDoc, opts.NewDocID, "after pushed quartzo rebuild"); err != nil {
		return RebuildQuartzoResult{}, err
	}
	rootDelta, err := engine.EncodeDelta(rootDoc, rootStateVector)
	if err != nil {
		return RebuildQuartzoResult{}, err
	}
	if err := client.PushDocUpdate(opts.WorkspaceID, opts.WorkspaceID, rootDelta); err != nil {
		return RebuildQuartzoResult{}, fmt.Errorf("push root registration: %w", err)
	}
	return result, nil
}

func writeRebuildSnapshots(dir, workspaceID, docID, title, docB64, rootB64 string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	docRaw, err := base64.StdEncoding.DecodeString(docB64)
	if err != nil {
		return err
	}
	rootRaw, err := base64.StdEncoding.DecodeString(rootB64)
	if err != nil {
		return err
	}
	files := map[string][]byte{
		filepath.Join(dir, docID+".doc.bin"):        docRaw,
		filepath.Join(dir, docID+".doc.b64"):        []byte(docB64),
		filepath.Join(dir, workspaceID+".root.bin"): rootRaw,
		filepath.Join(dir, workspaceID+".root.b64"): []byte(rootB64),
		filepath.Join(dir, "manifest.json"): mustJSON(map[string]any{
			"workspace_id": workspaceID,
			"doc_id":       docID,
			"title":        title,
			"doc_bytes":    len(docRaw),
			"root_bytes":   len(rootRaw),
		}),
	}
	for path, data := range files {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			return err
		}
	}
	return nil
}

func mustJSON(v any) []byte {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return []byte("{}")
	}
	return raw
}

func waitForDoc(client *socketio.Client, workspaceID, docID string) error {
	var lastErr error
	for i := 0; i < 10; i++ {
		if i > 0 {
			time.Sleep(time.Duration(i) * 500 * time.Millisecond)
		}
		loaded, err := client.LoadDoc(workspaceID, docID)
		if err == nil && loaded != nil && loaded.Missing != "" {
			return nil
		}
		if err != nil {
			lastErr = err
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("doc %s did not become readable", docID)
}

func connect(cfg *config.Config, workspaceID string) (*socketio.Client, error) {
	bearer := cfg.AffineToken
	if bearer == "" {
		bearer = strings.TrimPrefix(cfg.AuthHeader(), "Bearer ")
	}
	client, err := socketio.Connect(socketio.WSURLFromGraphQL(strings.TrimRight(cfg.BaseURL, "/")+"/graphql"), "", bearer)
	if err != nil {
		return nil, err
	}
	if err := joinWorkspace(client, workspaceID); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

func rebuildDocScript(doc int, title string) string {
	return fmt.Sprintf(`
		(function() {
			var doc = globalThis._docs[%d];
			var blocks = doc.getMap("blocks");
			var page = null;
			var counts = {blocks: 0, notes: 0, images: 0, groups: {}};
			blocks.forEach(function(block) {
				counts.blocks++;
				if (!(block instanceof Y.Map)) return;
				var flavour = block.get("sys:flavour");
				if (flavour === "affine:page") page = block;
				if (flavour === "affine:image") counts.images++;
			});
			if (!page) throw new Error("affine:page block not found");
			page.set("prop:title", %s);

			function textOf(block) {
				var text = block && block.get && block.get("prop:text");
				return text && text.toString ? text.toString().trim() : "";
			}
			function firstText(note) {
				var children = note.get("sys:children");
				if (!children || !children.toArray) return "";
				var ids = children.toArray();
				for (var i = 0; i < ids.length; i++) {
					var s = textOf(blocks.get(ids[i]));
					if (s) return s;
				}
				return "";
			}
			function groupFor(name) {
				var s = (name || "").toLowerCase();
				if (/comercial|prospec|lead|sales|crm|cliente|pipeline|outbound/.test(s)) return "Comercial";
				if (/comunidade|carta clube|coding|criacao|criação|execucao|execução|qualidades|biblioteca/.test(s)) return "Comunidade";
				if (/infra|hetzner|contabo|traefik|coolify|infisical|uptime|lightrag|ragflow|papermark|documenso|penpot|chatbot/.test(s)) return "Infraestrutura";
				if (/trabalho|onboarding|curriculo|currículo|skill|brain|terminal|setup/.test(s)) return "Trabalho";
				if (/conte[uú]do|m[ií]dia|produto|app|showrunner|world|quartzo ai/.test(s)) return "Produto e Conteúdo";
				return "Ecossistema";
			}
			function parseXYWH(xywh) {
				try {
					var arr = JSON.parse(xywh || "[0,0,720,520]");
					return Array.isArray(arr) ? arr : [0,0,720,520];
				} catch (_) {
					return [0,0,720,520];
				}
			}
			function clamp(n, min, max) { return Math.max(min, Math.min(max, n)); }
			function isHeader(name) {
				return /^[A-ZÀ-Ú0-9\s-]{4,}$/.test((name || "").trim());
			}

			var children = page.get("sys:children");
			var groups = {
				"Ecossistema": [],
				"Trabalho": [],
				"Comercial": [],
				"Comunidade": [],
				"Infraestrutura": [],
				"Produto e Conteúdo": []
			};
			children.toArray().forEach(function(id, idx) {
				var block = blocks.get(id);
				if (!(block instanceof Y.Map) || block.get("sys:flavour") !== "affine:note") return;
				var name = firstText(block);
				var group = groupFor(name);
				groups[group].push({id: id, block: block, name: name, idx: idx});
			});

			var order = ["Ecossistema", "Trabalho", "Comercial", "Comunidade", "Infraestrutura", "Produto e Conteúdo"];
			var y = 0;
			for (var gi = 0; gi < order.length; gi++) {
				var groupName = order[gi];
				var notes = groups[groupName];
				if (!notes.length) continue;
				counts.groups[groupName] = notes.length;
				var x = 0;
				var rowH = 0;
				for (var i = 0; i < notes.length; i++) {
					var item = notes[i];
					var b = item.block;
					b.set("prop:displayMode", "both");
					var old = parseXYWH(b.get("prop:xywh"));
					var header = isHeader(item.name);
					var w = header ? 2400 : clamp(old[2] || 760, 620, 980);
					var h = header ? 220 : clamp(old[3] || 720, 440, 1120);
					if (!header && x + w > 5400) {
						x = 0;
						y += rowH + 180;
						rowH = 0;
					}
					b.set("prop:xywh", JSON.stringify([x, y, w, h]));
					x += w + 140;
					rowH = Math.max(rowH, h);
					counts.notes++;
				}
				y += rowH + 420;
			}
			return JSON.stringify(counts);
		})()
	`, doc, strconv.Quote(title))
}

func registerRootDocScript(doc int, docID, title string) string {
	now := time.Now().UnixMilli()
	return fmt.Sprintf(`
		(function() {
			var doc = globalThis._docs[%d];
			var meta = doc.getMap("meta");
			var pages = meta.get("pages");
			if (!(pages instanceof Y.Array)) {
				pages = new Y.Array();
				meta.set("pages", pages);
			}
			for (var i = pages.length - 1; i >= 0; i--) {
				var page = pages.get(i);
				if (page && page.get && page.get("id") === %s) {
					pages.delete(i, 1);
				}
			}
			var entry = new Y.Map();
			entry.set("id", %s);
			entry.set("title", %s);
			entry.set("createDate", %d);
			entry.set("updatedDate", %d);
			entry.set("tags", []);
			pages.push([entry]);
			return JSON.stringify({pageCount: pages.length});
		})()
	`, doc, strconv.Quote(docID), strconv.Quote(docID), strconv.Quote(title), now, now)
}
