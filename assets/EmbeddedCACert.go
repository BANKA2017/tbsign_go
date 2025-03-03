package assets

import (
	"embed"
)

//go:embed ca/*
var EmbeddedCACert embed.FS
