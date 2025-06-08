package options

// ExpireOptions represents options for the EXPIRE command
type ExpireOptions struct {
	*Options
}

func NewExpireOptions() *ExpireOptions {
	opts := &ExpireOptions{
		Options: NewOptions(),
	}

	opts.RegisterOption("NX", "Set expiry only if the key has no expiry", []string{"XX", "GT", "LT"})
	opts.RegisterOption("XX", "Set expiry only if the key has an existing expiry", []string{"NX", "GT", "LT"})
	opts.RegisterOption("GT", "Set expiry only if the new expiry is greater than current one", []string{"NX", "XX", "LT"})
	opts.RegisterOption("LT", "Set expiry only if the new expiry is less than current one", []string{"NX", "XX", "GT"})

	return opts
}

func (o *ExpireOptions) IsNX() bool {
	return o.IsSet("NX")
}

func (o *ExpireOptions) IsXX() bool {
	return o.IsSet("XX")
}

func (o *ExpireOptions) IsGT() bool {
	return o.IsSet("GT")
}

func (o *ExpireOptions) IsLT() bool {
	return o.IsSet("LT")
}
