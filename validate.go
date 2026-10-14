package raml

import (
	"fmt"

	"github.com/acronis/go-stacktrace"
)

func (r *RAML) unwrapShape(shape *BaseShape, unwrapCache map[int64]*BaseShape) (*BaseShape, *stacktrace.StackTrace) {
	if !shape.unwrapped {
		shape = shape.CloneDetached()
		us, err := r.UnwrapShape(shape)
		if err != nil {
			return nil, StacktraceNewWrapped("unwrap shape", err, shape.Location,
				stacktrace.WithPosition(&shape.KeyPos),
				stacktrace.WithType(StacktraceTypeValidating))
		}
		_, err = r.FindAndMarkRecursion(us)
		if err != nil {
			return nil, StacktraceNewWrapped("find recursion", err, shape.Location,
				stacktrace.WithPosition(&shape.KeyPos),
				stacktrace.WithType(StacktraceTypeValidating))
		}
		unwrapCache[shape.ID] = us
		shape = us
	}
	return shape, nil
}

func (r *RAML) validateTypes(unwrapCache map[int64]*BaseShape) *stacktrace.StackTrace {
	var acc stacktrace.Accumulator
	for _, shapes := range r.fragmentTypeDefinitions {
		for _, shape := range shapes {
			shape, se := r.unwrapShape(shape, unwrapCache)
			if se != nil {
				acc.Add(se)
				continue
			}
			if err := shape.Check(); err != nil {
				acc.Add(StacktraceNewWrapped("check type", err, shape.Location,
					stacktrace.WithPosition(&shape.KeyPos),
					stacktrace.WithType(StacktraceTypeValidating)))
				continue
			}
			if err := r.validateShapeCommons(shape); err != nil {
				acc.Add(StacktraceNewWrapped("validate shape commons", err, shape.Location,
					stacktrace.WithPosition(&shape.KeyPos),
					stacktrace.WithType(StacktraceTypeValidating)))
				continue
			}
		}
	}
	return acc.Result()
}

func (r *RAML) validateDomainExtensions(unwrapCache map[int64]*BaseShape) *stacktrace.StackTrace {
	var acc stacktrace.Accumulator
	for _, item := range r.domainExtensions {
		db := item.DefinedBy
		if db == nil {
			continue
		}
		if !db.unwrapped {
			us, ok := unwrapCache[db.ID]
			if !ok {
				acc.Add(StacktraceNew("unwrapped shape not found", db.Location,
					stacktrace.WithPosition(&db.KeyPos),
					stacktrace.WithType(StacktraceTypeValidating)))
				continue
			}
			db = us
		}
		if err := db.Validate(item.Extension.Value.Raw); err != nil {
			// Point at the property the error is about, not at the start of
			// the whole annotation value.
			pos := item.Extension.errorPosition(err)
			acc.Add(StacktraceNewWrapped("check domain extension", err, item.Extension.Location,
				stacktrace.WithPosition(&pos),
				stacktrace.WithType(StacktraceTypeValidating)))
			continue
		}
	}

	return acc.Result()
}

func (r *RAML) ValidateShapes() error {
	// Unwrap cache stores the mapping of original IDs to unwrapped shapes
	// to ensure the original references (aliases and links) match.
	unwrapCache := make(map[int64]*BaseShape)

	var acc stacktrace.Accumulator
	acc.Add(r.validateTypes(unwrapCache))
	acc.Add(r.validateDomainExtensions(unwrapCache))
	if st := acc.Result(); st != nil {
		return st
	}
	return nil
}

func (r *RAML) validateObjectShape(s *ObjectShape) error {
	for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
		base := pair.Value.Base
		if err := r.validateShapeCommons(base); err != nil {
			return StacktraceNewWrapped("validate property", err, base.Location,
				stacktrace.WithPosition(&base.KeyPos), stacktrace.WithInfo("property", pair.Key))
		}
	}
	for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
		base := pair.Value.Base
		if err := r.validateShapeCommons(base); err != nil {
			return StacktraceNewWrapped("validate pattern property", err, base.Location,
				stacktrace.WithPosition(&base.KeyPos), stacktrace.WithInfo("property", pair.Key))
		}
	}
	return nil
}

func (r *RAML) validateShapeCommons(s *BaseShape) error {
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
					stacktrace.WithPosition(&shape.Base().KeyPos))
			}
		}
	case *UnionShape:
		for _, item := range shape.AnyOf {
			if err := r.validateShapeCommons(item); err != nil {
				return StacktraceNewWrapped("validate union item", err, shape.Base().Location,
					stacktrace.WithPosition(&shape.Base().KeyPos))
			}
		}
	}

	// Validate trait definition shapes
	for pair := s.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		facetDef := pair.Value
		if err := r.validateShapeCommons(facetDef.Base); err != nil {
			return StacktraceNewWrapped("validate custom facet definition", err, facetDef.Base.Location,
				stacktrace.WithPosition(&facetDef.Base.KeyPos), stacktrace.WithInfo("facet", pair.Key))
		}
	}

	return nil
}

func (r *RAML) validateExamples(base *BaseShape) error {
	if base.Example != nil {
		if base.Example.Strict == nil || base.Example.Strict.Value {
			if err := base.Validate(base.Example.Data.Value.Raw); err != nil {
				return StacktraceNewWrapped("validate example", err, base.Example.Location,
					stacktrace.WithPosition(&base.Example.KeyPos))
			}
		}
	}
	if base.Examples != nil {
		for pair := base.Examples.Map.Oldest(); pair != nil; pair = pair.Next() {
			ex := pair.Value
			if ex.Strict != nil && !ex.Strict.Value {
				continue
			}
			if err := base.Validate(ex.Data.Value.Raw); err != nil {
				return StacktraceNewWrapped("validate example", err, ex.Location,
					stacktrace.WithPosition(&ex.KeyPos))
			}
		}
	}
	if base.Default != nil {
		if err := base.Validate(base.Default.Value.Raw); err != nil {
			return StacktraceNewWrapped("validate default", err, base.Default.Location,
				stacktrace.WithPosition(&base.Default.ValuePos))
		}
	}
	return nil
}

func (r *RAML) validateShapeFacets(base *BaseShape) error {
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
				return StacktraceNew("duplicate custom facet", f.Base.Location,
					stacktrace.WithPosition(&f.Base.KeyPos), stacktrace.WithInfo("facet", f.Name))
			}
			validationFacetDefs[f.Name] = f
		}
		inherits = parent.Inherits
	}

	// Validate all unknown facets against facet definitions
	shapeFacets := base.CustomShapeFacets
	for k, facetDef := range validationFacetDefs {
		f, ok := shapeFacets.Get(k)
		if !ok {
			if facetDef.Required {
				return StacktraceNew("required custom facet is missing", base.Location,
					stacktrace.WithPosition(&base.KeyPos), stacktrace.WithInfo("facet", k))
			}
			continue
		}
		if err := facetDef.Base.Validate(f.Value.Raw); err != nil {
			pos := f.errorPosition(err)
			return StacktraceNewWrapped("validate custom facet", err, f.Location,
				stacktrace.WithPosition(&pos), stacktrace.WithInfo("facet", k))
		}
	}

	// If we encounter an undefined facet - it's an error.
	for pair := shapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		k, f := pair.Key, pair.Value
		if _, ok := validationFacetDefs[k]; !ok {
			return StacktraceNew("unknown facet", f.Location, stacktrace.WithPosition(&f.KeyPos),
				stacktrace.WithInfo("facet", k))
		}
	}
	return nil
}
