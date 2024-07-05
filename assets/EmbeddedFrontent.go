package assets

import "embed"

//go:embed all:dist/*
var EmbeddedFrontent embed.FS
