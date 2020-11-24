package golden

import (
	"github.com/tommy351/goldga"
)

func Match() *goldga.Matcher {
	matcher := goldga.Match()
	matcher.Serializer = YAMLSerializer{}
	matcher.Transformer = ObjectTransformer{}

	return matcher
}
