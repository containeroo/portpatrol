package dynflags

type FlagType string

const (
	FlagTypeStringSlice FlagType = "..STRING"
	FlagTypeString      FlagType = "STRING"
	FlagTypeInt         FlagType = "INT"
	FlagTypeBool        FlagType = "BOOL"
	FlagTypeDuration    FlagType = "DURATION"
	FlagTypeFloat       FlagType = "FLOAT"
	FlagTypeURL         FlagType = "URL"
)

// Flag represents a single configuration flag
type Flag struct {
	Default interface{} // Default value for the flag
	Type    FlagType    // Type of the flag
	Usage   string      // Description for usage
	Value   FlagValue   // Encapsulated parsing and value-setting logic
}

// FlagValue interface encapsulates parsing and value-setting logic
type FlagValue interface {
	Parse(value string) (interface{}, error)
	Set(value interface{}) error
}
