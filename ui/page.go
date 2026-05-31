package ui

import "gbenson.net/go/zmachine/surface"

type Renderable interface {
	Render(r Renderer)
}

type Page interface {
	Renderable
}

type Updatable interface {
	Page
	Update(*surface.State)
}
