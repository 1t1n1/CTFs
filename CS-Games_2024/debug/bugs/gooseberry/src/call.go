package main

import (
	"fmt"
)

type Call Data

func (c Call) Eval(d *Data) error {
	if c.Type != "list" {
		return fmt.Errorf("expected list, got %s", c.Type)
	}

	if len(*c.ListValue) == 0 {
		return fmt.Errorf("expected at least 2 elements, got %d", len(*c.ListValue))
	}

	var f MappingFunction

	method := (*c.ListValue)[0]

	if method.Type != "string" {
		return fmt.Errorf("expected string, got %s value: %+v", method.Type, method.Value)
	}

	switch method.StringValue {
	case "add":
		args := (*c.ListValue)[1:]
		if len(args) != 1 {
			return fmt.Errorf("expected 1 argument, got %d", len(args))
		}
		f = func(d *Data) error {
			Add(d, args[0])
			return nil
		}
	case "inverse":
		f = func(d *Data) error {
			switch d.Type {
			case "int":
				d.IntValue = -d.IntValue
				d.Value = d.IntValue
			case "string":
				Reverse(([]rune)(d.StringValue))
				d.Value = d.StringValue
			}
			return nil
		}
	case "sum":
		total := 0
		f = func(d *Data) error {
			total += d.Value
			return nil
		}
		d.Map(f)
		*d = Data{
			Type:     "int",
			Value:    total,
		}
		return nil
	default:
		return fmt.Errorf("unknown method: %s %v", method.StringValue)
	}

	return d.Map(f)
}
