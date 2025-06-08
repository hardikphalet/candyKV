package options

import (
	"fmt"
	"strings"
)

type Option struct {
	Name         string
	Description  string
	Incompatible []string // List of option names that are incompatible with this option
}

type Options struct {
	options map[string]Option
	active  map[string]bool
}

func NewOptions() *Options {
	return &Options{
		options: make(map[string]Option),
		active:  make(map[string]bool),
	}
}

func (o *Options) RegisterOption(name, description string, incompatible []string) {
	o.options[name] = Option{
		Name:         name,
		Description:  description,
		Incompatible: incompatible,
	}
}

func (o *Options) Set(name string) error {
	name = strings.ToUpper(name)
	if _, exists := o.options[name]; !exists {
		return fmt.Errorf("unknown option: %s", name)
	}

	for activeOpt := range o.active {
		for _, incompatible := range o.options[name].Incompatible {
			if strings.ToUpper(incompatible) == activeOpt {
				return fmt.Errorf("option %s is incompatible with %s", name, activeOpt)
			}
		}
	}

	o.active[name] = true
	return nil
}

func (o *Options) IsSet(name string) bool {
	return o.active[strings.ToUpper(name)]
}

func (o *Options) Clear() {
	o.active = make(map[string]bool)
}

func (o *Options) GetActive() []string {
	active := make([]string, 0, len(o.active))
	for opt := range o.active {
		active = append(active, opt)
	}
	return active
}
