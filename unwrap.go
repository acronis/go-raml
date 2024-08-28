package raml

import "fmt"

func UnwrapShapes() error {
	for _, frag := range GetRegistry().fragmentsCache {
		lib, ok := frag.(*Library)
		if !ok {
			continue
		}
		for _, s := range lib.Types {
			us, err := UnwrapShape(s, true, true, make([]string, 0))
			if err != nil {
				return fmt.Errorf("resolve shape: %w", err)
			}
			*s = us
		}
	}

	return nil
}

func Inherit(source Shape, target Shape) (Shape, error) {
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
// Note that this method removes information about links and inheritance.
// NOTE: This function is not thread-safe. Use Clone() to create a copy of the shape before unwrapping if necessary.
func UnwrapShape(s *Shape, unwrapLinks bool, unwrapInherits bool, history []string) (Shape, error) {
	target := *s

	base := target.Base()
	sid := base.Id
	for _, item := range history {
		if item == base.Id {
			// TODO: Probably should insert RecursiveShape instead of target.
			return target, nil
		}
	}
	var source Shape
	link := base.Link
	inherits := base.Inherits
	if unwrapLinks && link != nil {
		us, err := UnwrapShape(link.Shape, unwrapLinks, unwrapInherits, append(history, sid))
		if err != nil {
			return nil, NewWrappedError("link unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		source = us

		base.Link = nil
	} else if unwrapInherits && len(inherits) > 0 {
		// TODO: Taking the first item is probably not a good idea, but it works.
		ss, err := UnwrapShape(inherits[0], unwrapLinks, unwrapInherits, append(history, sid))
		if err != nil {
			return nil, NewWrappedError("parent unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		for i := 1; i < len(inherits); i++ {
			us, err := UnwrapShape(inherits[i], unwrapLinks, unwrapInherits, append(history, sid))
			if err != nil {
				return nil, err
			}
			_, err = ss.Inherit(us)
			if err != nil {
				return nil, NewWrappedError("multiple parents unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
		source = ss

		base.Inherits = nil
	}

	if t, ok := target.(*ArrayShape); ok && t.Items != nil {
		_, err := UnwrapShape(t.Items, unwrapLinks, unwrapInherits, append(history, sid))
		if err != nil {
			return nil, NewWrappedError("array item unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
	} else if t, ok := target.(*ObjectShape); ok && t.Properties != nil {
		for _, prop := range t.Properties {
			_, err := UnwrapShape(prop.Shape, unwrapLinks, unwrapInherits, append(history, sid))
			if err != nil {
				return nil, NewWrappedError("object property unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
	} else if t, ok := target.(*UnionShape); ok && t.AnyOf != nil {
		for _, item := range t.AnyOf {
			_, err := UnwrapShape(item, unwrapLinks, unwrapInherits, append(history, sid))
			if err != nil {
				return nil, NewWrappedError("union unwrap", err, target.Base().Location, WithPosition(&target.Base().Position))
			}
		}
	}

	if source != nil {
		ms, err := Inherit(source, target)
		if err != nil {
			return nil, NewWrappedError("merge shapes", err, target.Base().Location, WithPosition(&target.Base().Position))
		}
		return ms, nil
	}
	return target, nil
}
