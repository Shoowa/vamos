package server

// Health offers summarized data.
type Health struct {
	Rdbms bool
}

func (h *Health) PassFail() bool {
	return h.Rdbms
}
