package assets

import (
	"embed"
)

//go:embed ca/*
var EmbeddedCACert embed.FS

//go:embed all:dist/*
var EmbeddedFrontend embed.FS

//go:embed sql/*
var EmbeddedSQL embed.FS

//go:embed upgrade/*
var EmbeddedUpgradeFiles embed.FS
