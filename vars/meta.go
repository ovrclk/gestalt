package vars

type Meta interface {
	Require(...string) Meta
	Export(...string) Meta
	Merge(Meta) Meta
}

func NewMeta() Meta {
	return &meta{}
}

type meta struct {
}

func (m *meta) Require(keys ...string) Meta {
	return m
}

func (m *meta) Export(keys ...string) Meta {
	return m
}

func (m *meta) Merge(other Meta) Meta {
	return m
}
