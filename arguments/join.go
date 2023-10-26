package arguments

import "github.com/grafana/grafana-build/pipeline"

func Join(f ...[]pipeline.Argument) []pipeline.Argument {
	r := []pipeline.Argument{}
	for _, v := range f {
		r = append(r, v...)
	}

	return r
}
