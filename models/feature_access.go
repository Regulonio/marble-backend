package models

type FeatureAccess int

const (
	Restricted FeatureAccess = iota
	Allowed
	Test
	MissingConfiguration
	UnknownFeatureAccess
)

var ValidFeaturesAccess = []FeatureAccess{Allowed, Restricted, Test}

// Provide a string value for each outcome
func (f FeatureAccess) String() string {
	switch f {
	case Allowed:
		return "allowed"
	case Restricted:
		return "restricted"
	case Test:
		return "test"
	case MissingConfiguration:
		return "missing_configuration"
	}
	return "unknown"
}

// Provide an Outcome from a string value
func FeatureAccessFrom(s string) FeatureAccess {
	switch s {
	case "allowed":
		return Allowed
	case "restricted":
		return Restricted
	case "test":
		return Test
	case "missing_configuration":
		return MissingConfiguration
	}
	return UnknownFeatureAccess
}

func (f FeatureAccess) IsAllowed() bool {
	return f == Allowed || f == Test
}
