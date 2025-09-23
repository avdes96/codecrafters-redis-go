package entry

type Stream struct {
	data map[string]map[string]string
}

func NewStream() *Stream {
	return &Stream{
		data: make(map[string]map[string]string),
	}
}

func (s *Stream) Add(id string, field string, value string) {
	if _, ok := s.data[id]; !ok {
		s.data[id] = make(map[string]string)
	}
	s.data[id][field] = value
}

func (s *Stream) Type() string {
	return "stream"
}
