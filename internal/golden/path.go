package golden

import "fmt"

func Path(name string) string {
	return fmt.Sprintf("testdata/%s.golden", name)
}
