package bindata

import "embed"

//go:embed static/*
var Fs embed.FS
