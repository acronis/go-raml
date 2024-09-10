package raml

func (r *RAML) ValidateShapes() error {
	// TODO: Shapes must be unwrapped before validation.
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			for _, shape := range f.AnnotationTypes {
				s := *shape
				if err := s.Check(); err != nil {
					return NewWrappedError("check annotation type", err, s.Base().Location, WithPosition(&s.Base().Position))
				}
				if err := r.validateShapeCommons(s); err != nil {
					return NewWrappedError("validate shape commons", err, s.Base().Location, WithPosition(&s.Base().Position))
				}
			}
			for _, shape := range f.Types {
				s := *shape
				if err := s.Check(); err != nil {
					return NewWrappedError("check type", err, s.Base().Location, WithPosition(&s.Base().Position))
				}
				if err := r.validateShapeCommons(s); err != nil {
					return NewWrappedError("validate shape commons", err, s.Base().Location, WithPosition(&s.Base().Position))
				}
			}
		case *DataType:
			s := *f.Shape
			if err := s.Check(); err != nil {
				return NewWrappedError("check data type", err, s.Base().Location, WithPosition(&s.Base().Position))
			}
			if err := r.validateShapeCommons(s); err != nil {
				return NewWrappedError("validate shape commons", err, s.Base().Location, WithPosition(&s.Base().Position))
			}
		}
	}
	for _, item := range r.domainExtensions {
		if err := (*item.DefinedBy).Validate(item.Extension.Value, "$"); err != nil {
			return NewWrappedError("check domain extension", err, item.Extension.Location, WithPosition(&item.Extension.Position))
		}
	}
	return nil
}

func (r *RAML) validateShapeCommons(s Shape) error {
	if err := r.validateShapeFacets(s); err != nil {
		return err
	}
	if err := r.validateExamples(s); err != nil {
		return err
	}

	switch s := s.(type) {
	case *ObjectShape:
		if s.Properties != nil {
			for k, prop := range s.Properties {
				s := *prop.Shape
				if err := r.validateShapeCommons(s); err != nil {
					return NewWrappedError("validate property", err, s.Base().Location, WithPosition(&s.Base().Position), WithInfo("property", k))
				}
			}
			for k, prop := range s.PatternProperties {
				s := *prop.Shape
				if err := r.validateShapeCommons(s); err != nil {
					return NewWrappedError("validate pattern property", err, s.Base().Location, WithPosition(&s.Base().Position), WithInfo("property", k))
				}
			}
		}
	case *ArrayShape:
		if s.Items != nil {
			if err := r.validateShapeCommons(*s.Items); err != nil {
				return NewWrappedError("validate items", err, s.Base().Location, WithPosition(&s.Base().Position))
			}
		}
	case *UnionShape:
		for _, item := range s.AnyOf {
			if err := r.validateShapeCommons(*item); err != nil {
				return NewWrappedError("validate union item", err, s.Base().Location, WithPosition(&s.Base().Position))
			}
		}
	}
	return nil
}

func (r *RAML) validateExamples(s Shape) error {
	base := s.Base()
	if base.Example != nil {
		if err := s.Validate(base.Example.Data.Value, "$"); err != nil {
			return NewWrappedError("validate example", err, base.Example.Location, WithPosition(&base.Example.Position))
		}
	}
	if base.Examples != nil {
		for _, ex := range base.Examples.Map {
			if err := s.Validate(ex.Data.Value, "$"); err != nil {
				return NewWrappedError("validate example", err, ex.Location, WithPosition(&ex.Position))
			}
		}
	}
	if base.Default != nil {
		if err := s.Validate(base.Default.Value, "$"); err != nil {
			return NewWrappedError("validate default", err, base.Default.Location, WithPosition(&base.Default.Position))
		}
	}
	return nil
}

func (r *RAML) validateShapeFacets(s Shape) error {
	// TODO: Doesn't support multiple inheritance.
	base := s.Base()
	inherits := base.Inherits
	shapeFacetDefs := base.CustomShapeFacetDefinitions
	validationFacetDefs := make(map[string]Property)
	for {
		if len(inherits) == 0 {
			break
		}
		parent := *inherits[0]
		for _, f := range parent.Base().CustomShapeFacetDefinitions {
			if _, ok := shapeFacetDefs[f.Name]; ok {
				base := (*f.Shape).Base()
				return NewError("duplicate custom facet", base.Location, WithPosition(&base.Position), WithInfo("facet", f.Name))
			}
			validationFacetDefs[f.Name] = f
		}
		inherits = parent.Base().Inherits
	}

	shapeFacets := base.CustomShapeFacets
	for k, facetDef := range validationFacetDefs {
		f, ok := shapeFacets[k]
		if !ok {
			if facetDef.Required {
				return NewError("required custom facet is missing", base.Location, WithPosition(&base.Position), WithInfo("facet", k))
			}
			continue
		}
		if err := (*facetDef.Shape).Validate(f.Value, "$"); err != nil {
			return NewWrappedError("validate custom facet", err, f.Location, WithPosition(&f.Position), WithInfo("facet", k))
		}
	}

	for k, f := range shapeFacets {
		if _, ok := validationFacetDefs[k]; !ok {
			return NewError("unknown facet", f.Location, WithPosition(&f.Position), WithInfo("facet", k))
		}
	}
	return nil
}
