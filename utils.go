package intertype

import "go/types"

func isIncluded(hay string, stack []string) bool {
	for i := range stack {
		if hay == stack[i] {
			return true
		}
	}
	return false
}

func checkImpossibleTypes(badTypes []string, dynamicTypes []types.Type) (extraTypes []string) {
	var dynTyps []string

	for i := range dynamicTypes {
		dynTyps = append(dynTyps, dynamicTypes[i].String())
	}

	for i := range dynTyps {
		if isIncluded(dynTyps[i], badTypes) {
			extraTypes = append(extraTypes, dynTyps[i])
		}
	}

	return extraTypes
}

func checkPossibleTypes(possibleTyps []string, dynamicTypes []types.Type) (missingTypes, impossibleTypes []string) {
	var dynTyps []string

	nilIncluded := false
	for i := range possibleTyps {
		if possibleTyps[i] == "untyped nil" {
			nilIncluded = true
		}
	}

	if !nilIncluded {
		possibleTyps = append(possibleTyps, "untyped nil")
	}

	for i := range dynamicTypes {
		dynTyps = append(dynTyps, dynamicTypes[i].String())
	}

	for i := range possibleTyps {
		if !isIncluded(possibleTyps[i], dynTyps) {
			missingTypes = append(missingTypes, possibleTyps[i])
		}
	}
	for i := range dynTyps {
		if !isIncluded(dynTyps[i], possibleTyps) {
			impossibleTypes = append(impossibleTypes, dynTyps[i])
		}
	}

	return missingTypes, impossibleTypes
}
