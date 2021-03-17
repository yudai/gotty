package server

import (
	"embed"
	_ "embed"
)

//go:embed static/*
var f embed.FS
