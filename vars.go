package gestalt

type Vars interface {
	Put(string, string)
	Get(string) string

	Count() int
}

type vars struct {
	values map[string]string
}

func NewVars() Vars {
	return &vars{make(map[string]string)}
}

func (v *vars) Put(key, val string) {
	v.values[key] = val
}

func (v *vars) Get(key string) string {
	return v.values[key]
}

func (v *vars) Count() int {
	return len(v.values)
}
