package vars

type Meta interface {
	Require(...string) Meta
	Export(...string) Meta
	Requires() []string
	Exports() []string
	Merge(Meta) Meta
}

type meta struct {
	exports  []string
	requires []string
}

func NewMeta() Meta {
	return &meta{}
}

func (m *meta) Requires() []string {
	return m.requires
}

func (m *meta) Exports() []string {
	return m.exports
}

func (m *meta) Require(keys ...string) Meta {
	m.requires = append(m.requires, keys...)
	return m
}

func (m *meta) Export(keys ...string) Meta {
	m.exports = append(m.exports, keys...)
	return m
}

func (m *meta) Merge(other Meta) Meta {
	return &meta{
		requires: append(m.requires, other.Requires()...),
		exports:  append(m.exports, other.Requires()...),
	}
}
