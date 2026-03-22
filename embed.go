package main

import "embed"

//go:embed static
var StaticFiles embed.FS

//go:embed mcp/docs/manifest.json mcp/docs/ultron
var DocsFiles embed.FS

//go:embed config.toml
var DefaultConfig []byte
