package bugsnag

const (
	// AuthTypeBasic is a basic auth.
	AuthTypeBasic AuthType = "basic"
	// AuthTypeToken is a token auth.
	AuthTypeToken AuthType = "token"
)

// AuthType is a bugsnag authentication type.
// Currently supports basic and token (PAT).
type AuthType string

// String implements stringer interface.
func (at AuthType) String() string {
	return string(at)
}
