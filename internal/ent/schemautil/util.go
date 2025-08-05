package schemautil

// EdgeAuthCheckName is a name for the edge auth check
var EdgeAuthCheckName = "EDGE_AUTH_CHECK"

// Name returns the name of the SchemaGenAnnotation
func (a EdgeAccessAnnotation) Name() string {
	return EdgeAuthCheckName
}

type EdgeAccessAnnotation struct {
	ObjectType        string
	RequiresEditCheck bool
}

func EdgeAuthCheck(objectType string) EdgeAccessAnnotation {
	return EdgeAccessAnnotation{
		ObjectType:        objectType,
		RequiresEditCheck: true,
	}
}

func EdgeNoAuthCheck(objectType string) EdgeAccessAnnotation {
	return EdgeAccessAnnotation{
		ObjectType:        objectType,
		RequiresEditCheck: false,
	}
}
