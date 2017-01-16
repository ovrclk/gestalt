package vars

type Vars interface {
	Put(string, string)
	Get(string) string
	Count() int

	Clone() Vars
}

type varmap struct {
	values map[string]string
}

func NewVars() *varmap {
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

func (v *varmap) Clone() Vars {
	clone := NewVars()
	for k, v := range v.values {
		clone.values[k] = v
	}
	return clone
}
