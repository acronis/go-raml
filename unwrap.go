package raml

import "fmt"

// UnwrapShapes unwraps all shapes in the RAML.
func (r *RAML) UnwrapShapes() error {
	for _, shape := range r.GetShapePtrs() {
		us, err := r.UnwrapShape(shape, true, true, make([]string, 0))
		if err != nil {
			return fmt.Errorf("unwrap shape: %w", err)
		}
		*shape = us
	}
	return nil
}

// InheritBase writes inheritable properties of sourceBase into targetBase.
// Modifies targetBase in-place.
func InheritBase(sourceBase *BaseShape, targetBase *BaseShape) {
	// TODO: Maybe needs a bool switch for flexibility?
	for k, item := range sourceBase.CustomShapeFacets {
		if _, ok := targetBase.CustomShapeFacets[k]; !ok {
			targetBase.CustomShapeFacets[k] = item
		}
	}
	for k, item := range sourceBase.CustomDomainProperties {
		if _, ok := targetBase.CustomDomainProperties[k]; !ok {
			targetBase.CustomDomainProperties[k] = item
		}
	}
	// TODO: CustomShapeFacetDefinitions are not inheritable in context of unwrapper. But maybe they can be inheritable in other context?
}

// Inherit merges source shape into target shape.
func (r *RAML) Inherit(source Shape, target Shape) (Shape, error) {
	InheritBase(source.Base(), target.Base())
	// If source type is any, return target as is
	if _, ok := source.(*AnyShape); ok {
		return target, nil
	}

	sourceUnion, isSourceUnion := source.(*UnionShape)
	targetUnion, isTargetUnion := target.(*UnionShape)

	if isSourceUnion && !isTargetUnion {
		var filtered []*Shape
		for _, item := range sourceUnion.AnyOf {
			i := *item
			// If at least one union member has any type, the whole union is considered as any type.
			if _, ok := i.(*AnyShape); ok {
				return target, nil
			}
			// TODO: Check type compatibility
			if i.Base().Type == target.Base().Type {
				// Clone is required since we create new union members from source members
				ms, err := target.Clone().Inherit(i)
				if err != nil {
					return nil, NewWrappedError("merge shapes", err, target.Base().Location, WithPosition(&target.Base().Position))
				}
				filtered = append(filtered, &ms)
			}
		}
		if len(filtered) == 0 {
			return nil, NewError("failed to find compatible union member", target.Base().Location, WithPosition(&target.Base().Position))
		}
		// If only one union member remains - simplify to target type
		if len(filtered) == 1 {
			return *filtered[0], nil
		}
		// Convert target to union
		target.Base().Type = TypeUnion
		return &UnionShape{
			BaseShape: *target.Base(),
			UnionFacets: UnionFacets{
				AnyOf: filtered,
			},
		}, nil
	} else if isTargetUnion && !isSourceUnion {
		for _, item := range targetUnion.AnyOf {
			// Merge will raise an error in case any of union members has incompatible type
			_, err := (*item).Inherit(source)
			if err != nil {
				return nil, NewWrappedError("merge shapes", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
		return targetUnion, nil
	} else {
		// Primitive + Primitive (homogenous types) = Same type
		ms, err := target.Inherit(source)
		if err != nil {
			return nil, NewWrappedError("merge shapes", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		return ms, nil
	}
}

// Recursively unwraps shape in-place.
// Note that this method removes information about links.
// NOTE: This function is not thread-safe. Use Clone() to create a copy of the shape before unwrapping if necessary.
func (r *RAML) UnwrapShape(s *Shape, unwrapLinks bool, unwrapInherits bool, history []string) (Shape, error) {
	if s == nil {
		return nil, fmt.Errorf("shape is nil")
	}
	target := *s

	// Skip already unwrapped shapes
	if target.Base().IsUnwrapped() {
		return target, nil
	}

	base := target.Base()
	sid := base.Id
	for _, item := range history {
		if item == base.Id {
			// TODO: Probably should insert RecursiveShape instead of target.
			target.Base().unwrapped = true
			return target, nil
		}
	}
	var source Shape
	link := base.Link
	inherits := base.Inherits
	if unwrapLinks && link != nil {
		us, err := r.UnwrapShape(link.Shape, unwrapLinks, unwrapInherits, append(history, sid))
		if err != nil {
			return nil, NewWrappedError("link unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		source = us

		base.Link = nil
	} else if unwrapInherits && len(inherits) > 0 {
		unwrappedInherits := make([]*Shape, len(inherits))
		// TODO: Fix multiple inheritance unwrap.
		// Multiple inheritance members must be checked for compatibility with each other before unwrapping.
		ss, err := r.UnwrapShape(inherits[0], unwrapLinks, unwrapInherits, append(history, sid))
		if err != nil {
			return nil, NewWrappedError("parent unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		unwrappedInherits[0] = &ss
		for i := 1; i < len(inherits); i++ {
			us, err := r.UnwrapShape(inherits[i], unwrapLinks, unwrapInherits, append(history, sid))
			if err != nil {
				return nil, err
			}
			unwrappedInherits[i] = &us
			_, err = ss.Inherit(us)
			if err != nil {
				return nil, NewWrappedError("multiple parents unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
		source = ss

		base.Inherits = unwrappedInherits
	}

	if t, ok := target.(*ArrayShape); ok && t.Items != nil {
		_, err := r.UnwrapShape(t.Items, unwrapLinks, unwrapInherits, append(history, sid))
		if err != nil {
			return nil, NewWrappedError("array item unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
	} else if t, ok := target.(*ObjectShape); ok && t.Properties != nil {
		for _, prop := range t.Properties {
			_, err := r.UnwrapShape(prop.Shape, unwrapLinks, unwrapInherits, append(history, sid))
			if err != nil {
				return nil, NewWrappedError("object property unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
	} else if t, ok := target.(*UnionShape); ok && t.AnyOf != nil {
		for _, item := range t.AnyOf {
			_, err := r.UnwrapShape(item, unwrapLinks, unwrapInherits, append(history, sid))
			if err != nil {
				return nil, NewWrappedError("union unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
	}

	if source != nil {
		ms, err := r.Inherit(source, target)
		if err != nil {
			return nil, NewWrappedError("merge shapes", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		ms.Base().unwrapped = true
		return ms, nil
	}
	target.Base().unwrapped = true
	return target, nil
}
