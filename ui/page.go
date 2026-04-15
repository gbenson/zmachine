package ui

type Renderable interface {
	Render(r *Renderer)
}

type Page interface {
	Renderable
}
