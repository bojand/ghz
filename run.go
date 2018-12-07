package ghz

// Run executes the test
func Run(proto, call, host string, options ...Option) (*Report, error) {
	c, err := newConfig(options...)

	if err != nil {
		return nil, err
	}

	r := newReport(proto, call, host, c)
	return r, nil
}
