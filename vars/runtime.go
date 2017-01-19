package vars

func ImportTo(m Meta, from Vars, to Vars) {
	for _, key := range m.Requires() {
		to.Put(key, from.Get(key))
	}
}

func ExportTo(m Meta, from Vars, to Vars) {
	for _, key := range m.Exports() {
		to.Put(key, from.Get(key))
	}
}
