package assets

import (
	"embed"
)

//go:embed sql/*
var EmbeddedSQL embed.FS
