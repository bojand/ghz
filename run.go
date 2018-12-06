package ghz

// Run executes the test
func Run(proto, call, host string, options ...Option) (*Report, error) {
	c := NewConfig()
	for _, option := range options {
		err := option(c)

		return nil, err
	}

	r := &Report{}
	r.init(proto, call, host, c)
	return r, nil
}
