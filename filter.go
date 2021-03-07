package rsql

func (p *RSQL) parseFilter(values map[string]string, params *Params) error {
	val, ok := values[p.FilterTag]
	if !ok || len(val) < 1 {
		return nil
	}

	var err error
	params.Filters, err = ParseFilter(val)
	return err
}
