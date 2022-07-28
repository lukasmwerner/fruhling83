package main

import (
	_ "embed"
	"html/template"
)

type Templates struct {
	LandingPage *template.Template
	BoardPage   *template.Template
}

//go:embed templates/index.tmpl.html
var landingP string

//go:embed templates/board.tmpl.html
var boardP string

func NewTemplates() *Templates {
	return &Templates{
		LandingPage: template.Must(template.New("landing").Parse(landingP)),
		BoardPage:   template.Must(template.New("board").Parse(boardP)),
	}
}
