package assets

import "embed"

//go:embed dist/*
var EmbeddedFrontent embed.FS
