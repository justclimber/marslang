package iterpereter

import (
	"aakimov/marslang/object"
	"fmt"
	"log"
)

var Builtins = map[string]*object.Builtin{
	"print": &object.Builtin{
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				log.Fatalf("wrong number of arguments. got=%d, want=1", len(args))
			}
			fmt.Println(args[0].Inspect())
			return nil
		},
	},
}
