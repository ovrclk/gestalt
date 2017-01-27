package vars

type Meta interface {
	Require(...string) Meta
	Export(...string) Meta
	Default(string, string) Meta

	Requires() []string
	Exports() []string
	Defaults() map[string]string
	Merge(Meta) Meta
}

type meta struct {
	exports  []string
	requires []string
	defaults map[string]string
}

func NewMeta() Meta {
	return &meta{defaults: make(map[string]string)}
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

func (m *meta) Default(key, val string) Meta {
	m.defaults[key] = val
	return m
}

func (m *meta) Defaults() map[string]string {
	return m.defaults
}

func (m *meta) Merge(other Meta) Meta {
	defaults := make(map[string]string)

	for k, v := range m.defaults {
		defaults[k] = v
	}
	for k, v := range other.Defaults() {
		defaults[k] = v
	}

	return &meta{
		requires: append(m.requires, other.Requires()...),
		exports:  append(m.exports, other.Exports()...),
		defaults: defaults,
	}
}
