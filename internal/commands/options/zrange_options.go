package options

import "fmt"

type ZRangeOptions struct {
	*Options
	RangeType string // "BYSCORE", "BYLEX", or "" for index-based range
	Rev       bool
	Limit     struct {
		Offset int
		Count  int
	}
	WithScores bool
}

func NewZRangeOptions() *ZRangeOptions {
	opts := &ZRangeOptions{
		Options: NewOptions(),
	}

	opts.RegisterOption("BYSCORE", "Return elements with scores between min and max", []string{"BYLEX"})
	opts.RegisterOption("BYLEX", "Return elements with lexicographical ordering", []string{"BYSCORE"})
	opts.RegisterOption("REV", "Reverse the order of returned elements", nil)
	opts.RegisterOption("WITHSCORES", "Return scores along with members", nil)

	return opts
}

func (o *ZRangeOptions) Set(option string) error {
	if err := o.Options.Set(option); err != nil {
		return err
	}

	switch option {
	case "WITHSCORES":
		o.WithScores = true
	case "REV":
		o.Rev = true
	}

	return nil
}

func (o *ZRangeOptions) IsByScore() bool {
	return o.RangeType == "BYSCORE"
}

func (o *ZRangeOptions) IsByLex() bool {
	return o.RangeType == "BYLEX"
}

func (o *ZRangeOptions) IsRev() bool {
	return o.Rev
}

func (o *ZRangeOptions) IsWithScores() bool {
	return o.WithScores
}

func (o *ZRangeOptions) SetRangeType(rangeType string) error {
	switch rangeType {
	case "BYSCORE", "BYLEX":
		o.RangeType = rangeType
		return nil
	default:
		return fmt.Errorf("invalid range type: %s", rangeType)
	}
}

func (o *ZRangeOptions) SetLimit(offset, count int) error {
	if offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	if count < 0 {
		return fmt.Errorf("count must be non-negative")
	}
	o.Limit.Offset = offset
	o.Limit.Count = count
	return nil
}
