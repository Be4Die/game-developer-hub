package assets

import _ "embed"

// AdminerCSS содержит кастомную темную тему Adminer в стиле GDH.
//
//go:embed adminer.css
var AdminerCSS []byte

// AdminerPasswordPlugin содержит плагин Adminer для предзаполнения пароля из URL.
//
//go:embed password.php
var AdminerPasswordPlugin []byte
