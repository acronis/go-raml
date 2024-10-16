package raml

import (
	"fmt"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func (r *RAML) unwrapShape(shape *BaseShape, unwrapCache map[int64]*BaseShape) (*BaseShape, *stacktrace.StackTrace) {
	if !shape.unwrapped {
		shape = shape.CloneDetached()
		us, err := r.UnwrapShape(shape)
		if err != nil {
			return nil, StacktraceNewWrapped("unwrap shape", err, shape.Location,
				stacktrace.WithPosition(&shape.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
		}
		_, err = r.FindAndMarkRecursion(us)
		if err != nil {
			return nil, StacktraceNewWrapped("find recursion", err, shape.Location,
				stacktrace.WithPosition(&shape.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
		}
		unwrapCache[shape.ID] = us
		shape = us
	}
	return shape, nil
}

const HookBeforeValidateTypes HookKey = "RAML.validateTypes"

func (r *RAML) validateTypes(
	types *orderedmap.OrderedMap[string, *BaseShape],
	unwrapCache map[int64]*BaseShape,
) *stacktrace.StackTrace {
	if err := r.callHooks(HookBeforeValidateTypes, types, unwrapCache); err != nil {
		return StacktraceNewWrapped("handle step", err, r.GetLocation())
	}
	var st *stacktrace.StackTrace
	for pair := types.Oldest(); pair != nil; pair = pair.Next() {
		shape, se := r.unwrapShape(pair.Value, unwrapCache)
		if se != nil {
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
		if err := shape.Check(); err != nil {
			se = StacktraceNewWrapped("check type", err, shape.Location,
				stacktrace.WithPosition(&shape.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
		if err := r.validateShapeCommons(shape); err != nil {
			se = StacktraceNewWrapped("validate shape commons", err, shape.Location,
				stacktrace.WithPosition(&shape.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
	}
	return st
}

const HookBeforeValidateLibrary HookKey = "RAML.validateLibrary"

func (r *RAML) validateLibrary(f *Library, unwrapCache map[int64]*BaseShape) *stacktrace.StackTrace {
	if err := r.callHooks(HookBeforeValidateLibrary, f, unwrapCache); err != nil {
		return StacktraceNewWrapped("handle step", err, f.Location)
	}
	st := r.validateTypes(f.AnnotationTypes, unwrapCache)

	if se := r.validateTypes(f.Types, unwrapCache); se != nil {
		if st == nil {
			st = se
		} else {
			st = st.Append(se)
		}
	}
	return st
}

const HookBeforeValidateDataType HookKey = "RAML.validateDataType"

func (r *RAML) validateDataType(f *DataType, unwrapCache map[int64]*BaseShape) *stacktrace.StackTrace {
	if err := r.callHooks(HookBeforeValidateDataType, f, unwrapCache); err != nil {
		return StacktraceNewWrapped("handle step", err, f.Location)
	}
	s := f.Shape
	if !s.unwrapped {
		s = s.CloneDetached()
		us, err := r.UnwrapShape(s)
		if err != nil {
			return StacktraceNewWrapped("unwrap shape", err, s.Location,
				stacktrace.WithPosition(&s.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
		}
		_, err = r.FindAndMarkRecursion(us)
		if err != nil {
			return StacktraceNewWrapped("find recursion", err, s.Location,
				stacktrace.WithPosition(&s.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
		}
		unwrapCache[s.ID] = us
		s = us
	}
	if err := s.Check(); err != nil {
		return StacktraceNewWrapped("check data type", err, s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithType(stacktrace.TypeValidating))
	}
	if err := r.validateShapeCommons(s); err != nil {
		return StacktraceNewWrapped("validate shape commons", err, s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithType(stacktrace.TypeValidating))
	}
	return nil
}

const HookBeforeValidateFragments HookKey = "RAML.validateFragments"

func (r *RAML) validateFragments(unwrapCache map[int64]*BaseShape) *stacktrace.StackTrace {
	if err := r.callHooks(HookBeforeValidateFragments, unwrapCache); err != nil {
		return StacktraceNewWrapped("handle step", err, r.GetLocation())
	}
	var st *stacktrace.StackTrace
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			if err := r.validateLibrary(f, unwrapCache); err != nil {
				if st == nil {
					st = err
				} else {
					st = st.Append(err)
				}
			}
		case *DataType:
			if err := r.validateDataType(f, unwrapCache); err != nil {
				if st == nil {
					st = err
				} else {
					st = st.Append(err)
				}
			}
		}
	}
	return st
}

const HookBeforeValidateDomainExtensions HookKey = "RAML.validateDomainExtensions"

func (r *RAML) validateDomainExtensions(unwrapCache map[int64]*BaseShape) *stacktrace.StackTrace {
	if err := r.callHooks(HookBeforeValidateDomainExtensions, unwrapCache); err != nil {
		return StacktraceNewWrapped("handle step", err, r.GetLocation())
	}
	var st *stacktrace.StackTrace
	for _, item := range r.domainExtensions {
		db := item.DefinedBy
		if !db.unwrapped {
			us, ok := unwrapCache[db.ID]
			if !ok {
				se := stacktrace.New("unwrapped shape not found", db.Location,
					stacktrace.WithPosition(&db.Position),
					stacktrace.WithType(stacktrace.TypeValidating))
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
				continue
			}
			db = us
		}
		if err := db.Validate(item.Extension.Value); err != nil {
			se := StacktraceNewWrapped("check domain extension", err, item.Extension.Location,
				stacktrace.WithPosition(&item.Extension.Position),
				stacktrace.WithType(stacktrace.TypeValidating))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
	}

	return st
}

const HookBeforeValidateShapes HookKey = "RAML.ValidateShapes"

func (r *RAML) ValidateShapes() error {
	if err := r.callHooks(HookBeforeValidateShapes); err != nil {
		return err
	}
	// Unwrap cache stores the mapping of original IDs to unwrapped shapes
	// to ensure the original references (aliases and links) match.
	unwrapCache := make(map[int64]*BaseShape)

	st := r.validateFragments(unwrapCache)

	if se := r.validateDomainExtensions(unwrapCache); se != nil {
		if st == nil {
			st = se
		} else {
			st = st.Append(se)
		}
	}

	if st != nil {
		return st
	}
	return nil
}

const HookBeforeValidateObjectShape HookKey = "RAML.validateObjectShape"

func (r *RAML) validateObjectShape(s *ObjectShape) error {
	if err := r.callHooks(HookBeforeValidateObjectShape, s); err != nil {
		return err
	}
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			base := pair.Value.Base
			if err := r.validateShapeCommons(base); err != nil {
				return StacktraceNewWrapped("validate property", err, base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithInfo("property", pair.Key))
			}
		}
	}
	if s.PatternProperties != nil {
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			base := pair.Value.Base
			if err := r.validateShapeCommons(base); err != nil {
				return StacktraceNewWrapped("validate pattern property", err, base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithInfo("property", pair.Key))
			}
		}
	}
	return nil
}

const HookBeforeValidateShapeCommons HookKey = "RAML.validateShapeCommons"

func (r *RAML) validateShapeCommons(s *BaseShape) error {
	if err := r.callHooks(HookBeforeValidateShapeCommons, s); err != nil {
		return err
	}
	if err := r.validateShapeFacets(s); err != nil {
		return err
	}
	if err := r.validateExamples(s); err != nil {
		return err
	}

	switch shape := s.Shape.(type) {
	case *ObjectShape:
		if err := r.validateObjectShape(shape); err != nil {
			return fmt.Errorf("validate object shape: %w", err)
		}
	case *ArrayShape:
		if shape.Items != nil {
			if err := r.validateShapeCommons(shape.Items); err != nil {
				return StacktraceNewWrapped("validate items", err, shape.Base().Location,
					stacktrace.WithPosition(&shape.Base().Position))
			}
		}
	case *UnionShape:
		for _, item := range shape.AnyOf {
			if err := r.validateShapeCommons(item); err != nil {
				return StacktraceNewWrapped("validate union item", err, shape.Base().Location,
					stacktrace.WithPosition(&shape.Base().Position))
			}
		}
	}
	return nil
}

const HookBeforeValidateExamples HookKey = "RAML.validateExamples"

func (r *RAML) validateExamples(base *BaseShape) error {
	if err := r.callHooks(HookBeforeValidateExamples, base); err != nil {
		return err
	}
	if base.Example != nil {
		if err := base.Validate(base.Example.Data.Value); err != nil {
			return StacktraceNewWrapped("validate example", err, base.Example.Location,
				stacktrace.WithPosition(&base.Example.Position))
		}
	}
	if base.Examples != nil {
		for pair := base.Examples.Map.Oldest(); pair != nil; pair = pair.Next() {
			ex := pair.Value
			if err := base.Validate(ex.Data.Value); err != nil {
				return StacktraceNewWrapped("validate example", err, ex.Location,
					stacktrace.WithPosition(&ex.Position))
			}
		}
	}
	if base.Default != nil {
		if err := base.Validate(base.Default.Value); err != nil {
			return StacktraceNewWrapped("validate default", err, base.Default.Location,
				stacktrace.WithPosition(&base.Default.Position))
		}
	}
	return nil
}

const HookBeforeValidateShapeFacets HookKey = "RAML.validateShapeFacets"

func (r *RAML) validateShapeFacets(base *BaseShape) error {
	if err := r.callHooks(HookBeforeValidateShapeFacets, base); err != nil {
		return err
	}
	// TODO: Doesn't support multiple inheritance.
	inherits := base.Inherits
	shapeFacetDefs := base.CustomShapeFacetDefinitions
	validationFacetDefs := make(map[string]Property)
	for {
		if len(inherits) == 0 {
			break
		}
		parent := inherits[0]
		for pair := parent.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
			f := pair.Value
			if _, ok := shapeFacetDefs.Get(f.Name); ok {
				return stacktrace.New("duplicate custom facet", f.Base.Location,
					stacktrace.WithPosition(&f.Base.Position), stacktrace.WithInfo("facet", f.Name))
			}
			validationFacetDefs[f.Name] = f
		}
		inherits = parent.Inherits
	}

	shapeFacets := base.CustomShapeFacets
	for k, facetDef := range validationFacetDefs {
		f, ok := shapeFacets.Get(k)
		if !ok {
			if facetDef.Required {
				return stacktrace.New("required custom facet is missing", base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithInfo("facet", k))
			}
			continue
		}
		if err := facetDef.Base.Validate(f.Value); err != nil {
			return StacktraceNewWrapped("validate custom facet", err, f.Location,
				stacktrace.WithPosition(&f.Position), stacktrace.WithInfo("facet", k))
		}
	}

	for pair := shapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		k, f := pair.Key, pair.Value
		if _, ok := validationFacetDefs[k]; !ok {
			return stacktrace.New("unknown facet", f.Location, stacktrace.WithPosition(&f.Position),
				stacktrace.WithInfo("facet", k))
		}
	}
	return nil
}
