package gitutil

import "strings"

type RefType string

const (
	RefTypeTag    RefType = "tags"
	RefTypeBranch RefType = "heads"
)

type Ref struct {
	Type RefType
	Name string
}

func ParseRef(ref string) (Ref, bool) {
	chunks := strings.SplitN(ref, "/", 3)

	if len(chunks) != 3 || chunks[0] != "refs" {
		return Ref{}, false
	}

	switch RefType(chunks[1]) {
	case RefTypeBranch:
		return Ref{Type: RefTypeBranch, Name: chunks[2]}, true
	case RefTypeTag:
		return Ref{Type: RefTypeTag, Name: chunks[2]}, true
	}

	return Ref{}, false
}
