package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"affine-pp-cli/internal/config"
	"affine-pp-cli/internal/socketio"
	"affine-pp-cli/internal/yjs"
)

const workspaceID = "727cc066-a25e-4560-b68d-414b67cbc5c8"

func main() {
	cfg, err := config.Load("")
	must(err)
	bearer := cfg.AffineToken
	if bearer == "" {
		bearer = strings.TrimPrefix(cfg.AuthHeader(), "Bearer ")
	}
	client, err := socketio.Connect(socketio.WSURLFromGraphQL(strings.TrimRight(cfg.BaseURL, "/")+"/graphql"), "", bearer)
	must(err)
	defer client.Close()
	clientVersion := os.Getenv("AFFINE_WS_CLIENT_VERSION")
	if clientVersion == "" {
		clientVersion = "0.26.0"
	}
	must(client.JoinWorkspace(workspaceID, clientVersion))

	loaded, err := client.LoadDoc(workspaceID, workspaceID)
	must(err)
	engine, err := yjs.NewEngine()
	must(err)
	doc, err := engine.ApplyBase64Update(loaded.Missing)
	must(err)
	raw, err := engine.RunScript(fmt.Sprintf(`
		(function() {
			var doc = globalThis._docs[%d];
			var meta = doc.getMap("meta");
			var pages = meta.get("pages");
			var result = {metaKeys: [], pageCount: 0, pages: []};
			meta.forEach(function(v, k) { result.metaKeys.push(k); });
			if (pages && pages.toArray) {
				result.pageCount = pages.length;
				pages.toArray().slice(0, 20).forEach(function(page) {
					var o = {};
					if (page && page.forEach) page.forEach(function(v, k) { o[k] = v; });
					result.pages.push(o);
				});
			}
			return JSON.stringify(result);
		})()
	`, doc))
	must(err)
	var pretty any
	must(json.Unmarshal([]byte(raw), &pretty))
	out, err := json.MarshalIndent(pretty, "", "  ")
	must(err)
	fmt.Println(string(out))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
