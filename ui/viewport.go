package ui

type viewport struct {
	width  int
	height int
}

func (v *viewport) dim(w, h int) {
	v.width = w
	v.height = h
}
