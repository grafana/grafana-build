package pipelines

import (
	"fmt"
	"os"
)

func WriteToStdout(v []string) {
	for _, v := range v {
		fmt.Fprintln(os.Stdout, v)
	}
}
