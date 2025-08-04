package main

import _ "embed"

//go:embed resource/dark.css
var fileDarkCSS string

//go:embed resource/glightbox.min.css
var fileGLightboxCSS string

//go:embed resource/style.css
var fileStyleCSS string

//go:embed resource/glightbox.min.js
var fileGLightboxJS string

//go:embed resource/script.js
var fileScriptJS string
