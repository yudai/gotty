package server

import (
	"embed"
)

//go:embed static/*
var f embed.FS
