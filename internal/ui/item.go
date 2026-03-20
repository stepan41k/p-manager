package ui

type Item string

func (i Item) FilterValue() string { return "" }

// Если вы используете DefaultDelegate, ему также нужны эти методы:
// func (i passwordItem) Title() string       { return i.title }
// func (i passwordItem) Description() string { return i.username }

