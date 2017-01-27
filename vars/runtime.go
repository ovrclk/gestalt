package vars

func ImportTo(m Meta, from Vars, to Vars) {
	to.Merge(from)

	for k, v := range m.Defaults() {
		if !to.Has(k) {
			to.Put(k, v)
		}
	}
}

func ExportTo(m Meta, from Vars, to Vars) {
	for _, key := range m.Exports() {
		if from.Has(key) {
			to.Put(key, from.Get(key))
		}
	}
}
