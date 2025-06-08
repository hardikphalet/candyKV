package options

type ZAddOptions struct {
	*Options
}

func NewZAddOptions() *ZAddOptions {
	opts := &ZAddOptions{
		Options: NewOptions(),
	}

	opts.RegisterOption("NX", "Only add new elements, don't update already existing elements", []string{"XX"})
	opts.RegisterOption("XX", "Only update elements that already exist, don't add new elements", []string{"NX"})
	opts.RegisterOption("GT", "Only update existing elements if the new score is greater than the current score", []string{"LT"})
	opts.RegisterOption("LT", "Only update existing elements if the new score is less than the current score", []string{"GT"})
	opts.RegisterOption("CH", "Modify the return value to return the number of changed elements instead of new elements", nil)
	opts.RegisterOption("INCR", "Increment the score of an element instead of setting it", []string{"NX", "XX", "GT", "LT"})

	return opts
}

func (o *ZAddOptions) IsNX() bool {
	return o.IsSet("NX")
}

func (o *ZAddOptions) IsXX() bool {
	return o.IsSet("XX")
}

func (o *ZAddOptions) IsGT() bool {
	return o.IsSet("GT")
}

func (o *ZAddOptions) IsLT() bool {
	return o.IsSet("LT")
}

func (o *ZAddOptions) IsCH() bool {
	return o.IsSet("CH")
}

func (o *ZAddOptions) IsINCR() bool {
	return o.IsSet("INCR")
}
