package visitor

type visitResult struct {
	Method     string
	URL        string
	BodySize   int64
	StatusCode int
}

type visitResults []visitResult

func (r visitResults) Len() int {
	return len(r)
}

func (r visitResults) Less(i, j int) bool {
	return r[i].BodySize > r[j].BodySize // descending order
}

func (r visitResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
