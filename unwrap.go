package raml

import "fmt"

// UnwrapShapes unwraps all shapes in the RAML.
func (r *RAML) UnwrapShapes() error {
	// NOTE: With unwrap, we replace pointers to definitions instead of values by pointers to keep original parents unchanged.
	// Otherwise, parents will be also modified and unwrap may produce unpredictable results.
	// Unfortunately, this is required to properly support recursive shapes.
	// A more sophisticated approach is required to save memory and avoid copies.

	// We need to invalidate old cache and re-populate it because references will no longer be valid after unwrapping.
	r.fragmentTypes = make(map[string]map[string]*Shape)
	r.fragmentAnnotationTypes = make(map[string]map[string]*Shape)
	r.shapes = nil
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			for k, shape := range f.AnnotationTypes {
				us, err := r.UnwrapShape(shape, true, true, make([]Shape, 0))
				if err != nil {
					return fmt.Errorf("unwrap shape: %w", err)
				}
				ptr := &us
				f.AnnotationTypes[k] = ptr
				r.PutAnnotationTypeIntoFragment(us.Base().Name, f.Location, ptr)
				r.PutShapePtr(ptr)
			}
			for k, shape := range f.Types {
				us, err := r.UnwrapShape(shape, true, true, make([]Shape, 0))
				if err != nil {
					return fmt.Errorf("unwrap shape: %w", err)
				}
				ptr := &us
				f.Types[k] = ptr
				r.PutTypeIntoFragment(us.Base().Name, f.Location, ptr)
				r.PutShapePtr(ptr)
			}
		case *DataType:
			us, err := r.UnwrapShape(f.Shape, true, true, make([]Shape, 0))
			if err != nil {
				return fmt.Errorf("unwrap shape: %w", err)
			}
			ptr := &us
			f.Shape = ptr
			r.PutTypeIntoFragment(us.Base().Name, f.Location, ptr)
			r.PutShapePtr(ptr)
		}
	}
	// Links to definedBy must be updated after unwrapping.
	for _, item := range r.domainExtensions {
		db := *item.DefinedBy
		ptr, err := r.GetAnnotationTypeFromFragmentPtr(db.Base().Location, db.Base().Name)
		if err != nil {
			return fmt.Errorf("get from fragment: %w", err)
		}
		item.DefinedBy = ptr
	}
	return nil
}

// InheritBase writes inheritable properties of sourceBase into targetBase.
// Modifies targetBase in-place.
func (r *RAML) inheritBase(sourceBase *BaseShape, targetBase *BaseShape) {
	// Back-propagate shape ID from child to parent
	// to keep the original ID when inheriting properties.
	sourceBase.Id = targetBase.Id
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
	r.inheritBase(source.Base(), target.Base())
	// If source is recursive, return source as is
	if _, ok := source.(*RecursiveShape); ok {
		return source, nil
	}
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
				// Deep copy with ID change is required since we create new union members from source members
				cs := target.Clone()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.Base().Id = generateShapeId()
				ms, err := cs.Inherit(i)
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
		// Homogenous types produce same type
		ms, err := target.Inherit(source)
		if err != nil {
			return nil, NewWrappedError("merge shapes", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		return ms, nil
	}
}

// Recursively copies and unwraps a shape.
// Note that this method removes information about links.
func (r *RAML) UnwrapShape(s *Shape, unwrapLinks bool, unwrapInherits bool, history []Shape) (Shape, error) {
	if s == nil {
		return nil, fmt.Errorf("shape is nil")
	}
	// Perform deep copy to avoid modifying the original shape
	target := (*s).Clone()

	// Skip already unwrapped shapes
	if target.Base().IsUnwrapped() {
		return target, nil
	}

	base := target.Base()
	for _, item := range history {
		if item.Base().Id == base.Id {
			base.Inherits = nil
			return &RecursiveShape{BaseShape: *base, Head: &item}, nil
		}
	}
	history = append(history, target)

	var source Shape
	link := base.Link
	inherits := base.Inherits
	if unwrapLinks && link != nil {
		us, err := r.UnwrapShape(link.Shape, unwrapLinks, unwrapInherits, history)
		if err != nil {
			return nil, NewWrappedError("link unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		source = us

		base.Link = nil
	} else if unwrapInherits && len(inherits) > 0 {
		unwrappedInherits := make([]*Shape, len(inherits))
		// TODO: Fix multiple inheritance unwrap.
		// Multiple inheritance members must be checked for compatibility with each other before unwrapping.
		ss, err := r.UnwrapShape(inherits[0], unwrapLinks, unwrapInherits, history)
		if err != nil {
			return nil, NewWrappedError("parent unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		unwrappedInherits[0] = &ss
		for i := 1; i < len(inherits); i++ {
			us, err := r.UnwrapShape(inherits[i], unwrapLinks, unwrapInherits, history)
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
		us, err := r.UnwrapShape(t.Items, unwrapLinks, unwrapInherits, history)
		if err != nil {
			return nil, NewWrappedError("array item unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		*t.Items = us
	} else if t, ok := target.(*ObjectShape); ok && t.Properties != nil {
		for _, prop := range t.Properties {
			us, err := r.UnwrapShape(prop.Shape, unwrapLinks, unwrapInherits, history)
			if err != nil {
				return nil, NewWrappedError("object property unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
			*prop.Shape = us
		}
	} else if t, ok := target.(*UnionShape); ok {
		for _, item := range t.AnyOf {
			us, err := r.UnwrapShape(item, unwrapLinks, unwrapInherits, history)
			if err != nil {
				return nil, NewWrappedError("union unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
			*item = us
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
