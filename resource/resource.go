package resource

import _ "embed"

//go:embed common/dark.css
var FileDarkCSS string

//go:embed common/glightbox.min.css
var FileGLightboxCSS string

//go:embed common/style.css
var FileStyleCSS string

//go:embed common/glightbox.min.js
var FileGLightboxJS string

//go:embed common/script.js
var FileScriptJS string
