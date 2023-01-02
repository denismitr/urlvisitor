package visitor

type response struct {
	Method     string
	URL        string
	BodySize   int
	StatusCode int
}

type responses []response

func (r responses) Len() int {
	return len(r)
}

func (r responses) Less(i, j int) bool {
	return r[i].BodySize > r[j].BodySize // descending order
}

func (r responses) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
