package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

type parameter struct {
	WellKnownPath string
}

func main() {
	var params parameter
	protogen.Options{
		ParamFunc: func(name, value string) error {
			if name == "well_known" {
				params.WellKnownPath = value
				return nil
			}
			return fmt.Errorf("Unrecognized parameter: %s", name)
		},
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f, params)
		}
		return nil
	})
}
