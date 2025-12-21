package main

import (
	"encoding/json"
	"fmt"
)

type Data struct {
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	StringValue string      `json:"value"`
	ListValue   *[]*Data    `json:"value"`
	IntValue    int         `json:"value"`
}

func (d *Data) UnmarshalJSON(b []byte) error {
	type Alias Data
	aux := &struct {
		Value []byte `json:"value"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	switch d.Type {
	case "string":
		d.StringValue = string(aux.Value)
		return json.Unmarshal(aux.Value, &d.Value)
	case "int":
		d.IntValue = int(aux.Value[0])
		d.Value = d.IntValue

	case "list":
		var list []*Data
		if err := json.Unmarshal(aux.Value, &list); err != nil {
			return err
		}
		d.ListValue = &list
		d.Value = d.ListValue
	default:
		return fmt.Errorf("unknown type: %s", d.Type)
	}
	return nil
}

func (a Data) Add(b Data) {
	if a.Type == b.Type {
		switch a.Type {
		case "int":
			a.IntValue += b.IntValue
			a.Value = a.IntValue
		case "string":
			a.StringValue += b.StringValue
			a.Value = a.StringValue
		}
	}
}

type MappingFunction func(*Data) error

func (d *Data) Map(f MappingFunction) error {
	switch d.Type {
	case "list":
		for _, v := range *d.ListValue {
			err := v.Map(f)
			if err != nil {
				return err
			}
		}
	default:
		return f(d)
	}

	return nil
}

func (d Data) ToJson() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

func Add(a *Data, b *Data) error {
	a.Add(b)
	return nil
}
