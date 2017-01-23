package vars

func ImportTo(m Meta, from Vars, to Vars) {
	to.Merge(from)
}

func ExportTo(m Meta, from Vars, to Vars) {
	for _, key := range m.Exports() {
		if from.Has(key) {
			to.Put(key, from.Get(key))
		}
	}
}
