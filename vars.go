package gestalt

type Vars interface {
	Put(string, string)
	Get(string) string

	Count() int
}

type varmap struct {
	values map[string]string
}

func NewVars() Vars {
	return &varmap{make(map[string]string)}
}

func (v *varmap) Put(key, val string) {
	v.values[key] = val
}

func (v *varmap) Get(key string) string {
	return v.values[key]
}

func (v *varmap) Count() int {
	return len(v.values)
}
