package vars

type Vars interface {
	Put(string, string)
	Get(string) string
	Has(string) bool
	Unset(string)

	Keys() []string
	Count() int

	Merge(Vars) Vars
	Clone() Vars
}

type varmap struct {
	values map[string]string
}

func NewVars() Vars {
	return &varmap{make(map[string]string)}
}

func FromMap(values map[string]string) Vars {
	return &varmap{values: values}
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

func (v *varmap) Has(key string) bool {
	_, ok := v.values[key]
	return ok
}

func (v *varmap) Unset(key string) {
	delete(v.values, key)
}

func (v *varmap) Keys() []string {
	keys := make([]string, len(v.values))
	i := 0
	for k, _ := range v.values {
		keys[i] = k
		i++
	}
	return keys
}

func (v *varmap) Clone() Vars {
	clone := NewVars()
	for k, v := range v.values {
		clone.Put(k, v)
	}
	return clone
}

func (v *varmap) Merge(other Vars) Vars {
	for _, k := range other.Keys() {
		v.values[k] = other.Get(k)
	}
	return v
}
