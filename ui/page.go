package ui

type Renderable interface {
	Render(r Renderer)
}

type Page interface {
	Renderable
}

type Updatable interface {
	Page
	Update(deltas []float64, edges []Edge)
}
