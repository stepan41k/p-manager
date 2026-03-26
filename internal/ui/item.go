package ui

type Item struct {
	title string
	desc string
}

func (i Item) FilterValue() string { return i.title }
func (i Item) Title() string       { return i.title}
func (i Item) Description() string { return i.desc}

