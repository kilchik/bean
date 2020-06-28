package bear

type Note struct {
	Tag   string `meta:"tag"`
	Key   string `meta:"key"`
	Title string `meta:"title"`
	Text  []byte `meta:"text"`
}
