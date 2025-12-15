package raml

import (
	"fmt"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func (r *RAML) unwrapTypes() *stacktrace.StackTrace {
	var st *stacktrace.StackTrace
	for location, shapes := range r.fragmentTypeDefinitions {
		for _, shape := range shapes {
			if shape == nil {
				se := StacktraceNew("shape is nil", location,
					stacktrace.WithType(StacktraceTypeUnwrapping))
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
				continue
			}
			_, err := r.UnwrapShape(shape)
			if err != nil {
				se := StacktraceNewWrapped("unwrap shape", err, location,
					stacktrace.WithType(StacktraceTypeUnwrapping), stacktrace.WithPosition(&shape.Position))
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
				continue
			}
		}
	}
	return st
}

func (r *RAML) unwrapTraitDefinitions(traitsDefs *orderedmap.OrderedMap[string, *TraitDefinition]) {
	for pair := traitsDefs.Oldest(); pair != nil; pair = pair.Next() {
		trait := pair.Value
		if trait.Link != nil {
			traitsDefs.Set(pair.Key, trait.Link.Trait)
		}
	}
}

func (r *RAML) unwrapSecuritySchemeDefinitions(securitySchemeDefs *orderedmap.OrderedMap[string, *SecuritySchemeDefinition]) {
	for pair := securitySchemeDefs.Oldest(); pair != nil; pair = pair.Next() {
		securitySchemeDef := pair.Value
		if securitySchemeDef.Link != nil {
			securitySchemeDefs.Set(pair.Key, securitySchemeDef.Link.SecurityScheme)
		}
	}
}

func (r *RAML) unwrapAPIFragment(f *APIFragment) *stacktrace.StackTrace {
	r.unwrapTraitDefinitions(f.Traits)
	r.unwrapSecuritySchemeDefinitions(f.SecuritySchemes)
	st := r.applyTraits(f)
	se := r.applySecuritySchemes(f)
	if se != nil {
		if st == nil {
			st = se
		} else {
			st = st.Append(se)
		}
	}
	return st
}

func (r *RAML) unwrapLibrary(f *Library) *stacktrace.StackTrace {
	r.unwrapTraitDefinitions(f.Traits)
	r.unwrapSecuritySchemeDefinitions(f.SecuritySchemes)
	return nil
}

func (r *RAML) unwrapDataTypeFragment(f *DataTypeFragment) *stacktrace.StackTrace {
	if f.Shape == nil {
		return StacktraceNew("shape is nil", f.Location,
			stacktrace.WithType(StacktraceTypeUnwrapping))
	}
	_, err := r.UnwrapShape(f.Shape)
	if err != nil {
		return StacktraceNewWrapped("unwrap shape", err, f.Location,
			stacktrace.WithType(StacktraceTypeUnwrapping), stacktrace.WithPosition(&f.Shape.Position))
	}
	return nil
}

func (r *RAML) unwrapFragments() *stacktrace.StackTrace {
	st := r.unwrapTypes()
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *APIFragment:
			se := r.unwrapAPIFragment(f)
			if se != nil {
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
			}
		case *Library:
			se := r.unwrapLibrary(f)
			if se != nil {
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
			}
		case *DataTypeFragment:
			se := r.unwrapDataTypeFragment(f)
			if se != nil {
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
			}
		}
	}
	return st
}

func (r *RAML) applyTraits(api *APIFragment) *stacktrace.StackTrace {
	for _, endPoint := range r.endPoints {
		for _, trait := range endPoint.Traits {
			// Endpoint-level traits are propagated to operations
			for pair := endPoint.Operations.Oldest(); pair != nil; pair = pair.Next() {
				operation := pair.Value
				operation.Traits = append(operation.Traits, trait)
			}
		}
		for pair := endPoint.Operations.Oldest(); pair != nil; pair = pair.Next() {
			// Operation-level traits apply to a single operation
			operation := pair.Value
			for _, trait := range operation.Traits {
				if err := r.applyTrait(api, trait, operation); err != nil {
					return StacktraceNewWrapped("apply trait", err, trait.Location, stacktrace.WithPosition(&trait.Position),
						stacktrace.WithType(StacktraceTypeUnwrapping))
				}
			}
		}
	}
	return nil
}

func (r *RAML) applyTrait(api *APIFragment, trait *Trait, operation *Operation) *stacktrace.StackTrace {
	traitDef, err := api.GetTraitDefinition(trait.Name)
	if err != nil {
		return StacktraceNewWrapped("get trait definition", err, trait.Location,
			stacktrace.WithPosition(&trait.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
	}
	traitOperation, err := traitDef.compile(trait.Params)
	if err != nil {
		return StacktraceNewWrapped("compile trait", err, trait.Location,
			stacktrace.WithPosition(&trait.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
	}
	operation.merge(traitOperation)
	return nil
}

func (r *RAML) applySecuritySchemes(api *APIFragment) *stacktrace.StackTrace {
	for _, endPoint := range r.endPoints {
		for _, security := range endPoint.SecuredBy {
			// Endpoint-level security schemes are propagated to operations
			for pair := endPoint.Operations.Oldest(); pair != nil; pair = pair.Next() {
				operation := pair.Value
				operation.SecuredBy = append(operation.SecuredBy, security)
			}
		}
		for pair := endPoint.Operations.Oldest(); pair != nil; pair = pair.Next() {
			// Operation-level traits apply to a single operation
			operation := pair.Value
			for _, security := range operation.SecuredBy {
				if err := r.applySecurityScheme(api, security); err != nil {
					return StacktraceNewWrapped("apply security scheme", err, security.Location, stacktrace.WithPosition(&security.Position),
						stacktrace.WithType(StacktraceTypeUnwrapping))
				}
			}
		}
	}
	return nil
}

func (r *RAML) applySecurityScheme(api *APIFragment, securityScheme *SecurityScheme) *stacktrace.StackTrace {
	// TODO: Probably move to resolution stage
	// NOTE: If definition is already set, this is null security scheme.
	if securityScheme.Definition != nil {
		return nil
	}
	securitySchemeDef, err := api.GetSecuritySchemeDefinition(securityScheme.Name)
	if err != nil {
		return StacktraceNewWrapped("get security scheme definition", err, securityScheme.Location,
			stacktrace.WithPosition(&securityScheme.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
	}
	// TODO: Compile with parameters
	securityScheme.Definition = securitySchemeDef
	// NOTE: Security scheme description is not applied to the operation.
	// TODO: Maybe some checks can be implemented? Like method shadowing security status codes, headers and query parameters?
	return nil
}

func (r *RAML) unwrapDomainExtensions() *stacktrace.StackTrace {
	var st *stacktrace.StackTrace
	for _, item := range r.domainExtensions {
		db := item.DefinedBy
		ptr, err := r.GetAnnotationTypeFromFragmentPtr(db.Location, db.Name)
		if err != nil {
			se := StacktraceNewWrapped("get annotation from fragment", err, db.Location,
				stacktrace.WithPosition(&db.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
		item.DefinedBy = ptr
	}
	return st
}

// UnwrapShapes unwraps all shapes in the RAML in-place.
func (r *RAML) UnwrapShapes() error {
	// We need to invalidate old cache and re-populate it because references will no longer be valid after unwrapping.
	r.shapes = make([]*BaseShape, 0, len(r.shapes))
	if err := r.unwrapFragments(); err != nil {
		return fmt.Errorf("unwrap fragments: %w", err)
	}
	if err := r.markShapeRecursions(); err != nil {
		return fmt.Errorf("mark shape recursions: %w", err)
	}
	// Links to definedBy must be updated after unwrapping.
	if err := r.unwrapDomainExtensions(); err != nil {
		return fmt.Errorf("unwrap domain extensions: %w", err)
	}
	return nil
}

// markShapeRecursions marks recursive shapes by replacing the beginning of recursion with RecursiveShape in the RAML.
func (r *RAML) markShapeRecursions() error {
	// TODO: Maybe count shapes here?
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			for pair := f.AnnotationTypes.Oldest(); pair != nil; pair = pair.Next() {
				if _, err := r.FindAndMarkRecursion(pair.Value); err != nil {
					return err
				}
			}
			for pair := f.Types.Oldest(); pair != nil; pair = pair.Next() {
				if _, err := r.FindAndMarkRecursion(pair.Value); err != nil {
					return err
				}
			}
		case *DataTypeFragment:
			if _, err := r.FindAndMarkRecursion(f.Shape); err != nil {
				return err
			}
		}
	}
	return nil
}

const HookBeforeFindAndMarkRecursion HookKey = "RAML.FindAndMarkRecursion"

// FindAndMarkRecursion finds recursive shapes and replaces them with RecursiveShape.
func (r *RAML) FindAndMarkRecursion(base *BaseShape) (*BaseShape, error) {
	if err := r.callHooks(HookBeforeFindAndMarkRecursion, base); err != nil {
		return nil, err
	}
	if !base.IsUnwrapped() {
		return nil, fmt.Errorf("shape is not unwrapped")
	}

	if base.ShapeVisited {
		s := r.MakeRecursiveShape(base)
		s.unwrapped = true
		return s, nil
	}
	base.ShapeVisited = true

	var err error
	switch t := base.Shape.(type) {
	case *ArrayShape:
		err = r.findAndMarkRecursionInArrayShape(t)
	case *ObjectShape:
		err = r.findAndMarkRecursionInObjectShape(t)
	case *UnionShape:
		err = r.findAndMarkRecursionInUnionShape(t)
	}
	if err != nil {
		return nil, err
	}

	err = r.findAndMarkRecursionInCustomShapeFacetDefinitions(base)
	if err != nil {
		return nil, err
	}

	base.ShapeVisited = false
	return nil, ErrNil
}

func (r *RAML) findAndMarkRecursionInCustomShapeFacetDefinitions(base *BaseShape) error {
	for pair := base.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		rs, err := r.FindAndMarkRecursion(prop.Base)
		if err != nil {
			return fmt.Errorf("find and mark recursion: %w", err)
		}
		if rs != nil {
			prop.Base = rs
			base.CustomShapeFacetDefinitions.Set(pair.Key, prop)
		}
	}
	return nil
}

func (r *RAML) findAndMarkRecursionInArrayShape(t *ArrayShape) error {
	if t.Items != nil {
		rs, err := r.FindAndMarkRecursion(t.Items)
		if err != nil {
			return fmt.Errorf("find and mark recursion: %w", err)
		}
		if rs != nil {
			t.Items = rs
		}
	}
	return nil
}

func (r *RAML) findAndMarkRecursionInObjectShape(t *ObjectShape) error {
	if t.Properties != nil {
		for pair := t.Properties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			rs, err := r.FindAndMarkRecursion(prop.Base)
			if err != nil {
				return fmt.Errorf("find and mark recursion: %w", err)
			}
			if rs != nil {
				prop.Base = rs
				t.Properties.Set(pair.Key, prop)
			}
		}
	}
	if t.PatternProperties != nil {
		for pair := t.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			rs, err := r.FindAndMarkRecursion(prop.Base)
			if err != nil {
				return fmt.Errorf("find and mark recursion: %w", err)
			}
			if rs != nil {
				prop.Base = rs
				t.PatternProperties.Set(pair.Key, prop)
			}
		}
	}
	return nil
}

func (r *RAML) findAndMarkRecursionInUnionShape(t *UnionShape) error {
	for i, item := range t.AnyOf {
		rs, err := r.FindAndMarkRecursion(item)
		if err != nil {
			return fmt.Errorf("find and mark recursion: %w", err)
		}
		if rs != nil {
			t.AnyOf[i] = rs
		}
	}
	return nil
}

func (r *RAML) unwrapObjShape(objShape *ObjectShape) error {
	if objShape.Properties != nil {
		for pair := objShape.Properties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			us, err := r.UnwrapShape(prop.Base)
			if err != nil {
				return StacktraceNewWrapped("object property unwrap", err, objShape.Location,
					stacktrace.WithPosition(&objShape.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
			}
			prop.Base = us
			objShape.Properties.Set(pair.Key, prop)
		}
	}
	if objShape.PatternProperties != nil {
		for pair := objShape.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			us, err := r.UnwrapShape(prop.Base)
			if err != nil {
				return StacktraceNewWrapped("object pattern property unwrap", err, objShape.Location,
					stacktrace.WithPosition(&objShape.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
			}
			prop.Base = us
			objShape.PatternProperties.Set(pair.Key, prop)
		}
	}

	return nil
}

func (r *RAML) unwrapArrayShape(arrayShape *ArrayShape) error {
	if arrayShape.Items != nil {
		us, err := r.UnwrapShape(arrayShape.Items)
		if err != nil {
			return StacktraceNewWrapped("array item unwrap", err, arrayShape.Location,
				stacktrace.WithPosition(&arrayShape.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		arrayShape.Items = us
	}
	return nil
}

func (r *RAML) unwrapUnionShape(unionShape *UnionShape) error {
	for i, item := range unionShape.AnyOf {
		us, err := r.UnwrapShape(item)
		if err != nil {
			return StacktraceNewWrapped("union unwrap", err, unionShape.Location,
				stacktrace.WithPosition(&unionShape.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		unionShape.AnyOf[i] = us
	}
	return nil
}

func (r *RAML) unwrapParents(base *BaseShape) (*BaseShape, error) {
	var source *BaseShape
	switch {
	case base.Link != nil:
		us, err := r.UnwrapShape(base.Link.Shape)
		if err != nil {
			return nil, StacktraceNewWrapped("link unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		source = us
		base.Link = nil
	case len(base.Inherits) > 0:
		inherits := base.Inherits
		// FIXME: Multiple inheritance members must be checked for compatibility with each other before unwrapping.
		ss, err := r.UnwrapShape(inherits[0])
		if err != nil {
			return nil, StacktraceNewWrapped("parent unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		inherits[0] = ss
		for i := 1; i < len(inherits); i++ {
			us, errUnwrap := r.UnwrapShape(inherits[i])
			if errUnwrap != nil {
				return nil, errUnwrap
			}
			us, errUnwrap = ss.Inherit(us)
			if errUnwrap != nil {
				return nil, StacktraceNewWrapped("multiple parents unwrap", errUnwrap, base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
			}
			inherits[i] = us
		}
		source = ss
	}
	return source, nil
}

func (r *RAML) UnwrapTarget(target Shape) error {
	switch trg := target.(type) {
	case *ArrayShape:
		if err := r.unwrapArrayShape(trg); err != nil {
			return fmt.Errorf("unwrap array shape: %w", err)
		}
	case *ObjectShape:
		if err := r.unwrapObjShape(trg); err != nil {
			return fmt.Errorf("unwrap object shape: %w", err)
		}
	case *UnionShape:
		if err := r.unwrapUnionShape(trg); err != nil {
			return fmt.Errorf("unwrap union shape: %w", err)
		}
	}
	return nil
}

const HookBeforeUnwrapShape HookKey = "RAML.UnwrapShape"

// UnwrapShape recursively copies and unwraps a shape in-place. Use Clone() to create a copy of a shape if necessary.
// Note that this method removes information about links.
func (r *RAML) UnwrapShape(base *BaseShape) (*BaseShape, error) {
	if err := r.callHooks(HookBeforeUnwrapShape, base); err != nil {
		return nil, err
	}
	s := base.Shape
	if s == nil {
		return nil, fmt.Errorf("shape is nil")
	}

	// Skip already unwrapped shapes
	if base.IsUnwrapped() {
		return base, nil
	}
	base.SetUnwrapped()

	// NOTE: Type aliasing is not inheritance and is not used as a source. It must be unwrapped and returned as is.
	if base.Alias != nil {
		us, err := r.UnwrapShape(base.Alias)
		if err != nil {
			return nil, StacktraceNewWrapped("alias unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		r.PutShape(base)
		return base.AliasTo(us)
	}

	source, err := r.unwrapParents(base)
	if err != nil {
		return nil, StacktraceNewWrapped("unwrap parents", err, base.Location,
			stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
	}

	if err = r.UnwrapTarget(s); err != nil {
		return nil, StacktraceNewWrapped("unwrap target", err, base.Location,
			stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
	}

	for pair := base.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		us, errUnwrap := r.UnwrapShape(prop.Base)
		if errUnwrap != nil {
			return nil, StacktraceNewWrapped("custom shape facet definition unwrap", errUnwrap, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		prop.Base = us
		base.CustomShapeFacetDefinitions.Set(pair.Key, prop)
	}

	if source != nil {
		// Base shape inherits properties of the source shape in-place.
		is, errInherit := base.Inherit(source)
		if errInherit != nil {
			return nil, StacktraceNewWrapped("merge shapes", errInherit, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(StacktraceTypeUnwrapping))
		}
		base = is
	}
	r.PutShape(base)
	return base, nil
}
