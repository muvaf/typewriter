package scanner

import "go/types"

func WithReadCalls(s []string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.ReadOutputs = append(r.ReadOutputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithCreateCall(s string) Option {
	return func(r *RemoteCalls) {
		if t := r.scope.Lookup(s); t != nil {
			r.CreationInput = t.Type().(*types.Named)
		}
	}
}

func WithUpdateCalls(s []string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.UpdateInputs = append(r.UpdateInputs, t.Type().(*types.Named))
			}
		}
	}
}

func WithDeletionCalls(s []string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.DeletionInputs = append(r.DeletionInputs, t.Type().(*types.Named))
			}
		}
	}
}

type Option func(*RemoteCalls)

func NewRemoteCalls(s *types.Scope, opts ...Option) *RemoteCalls {
	r := &RemoteCalls{scope: s}
	for _, f := range opts {
		f(r)
	}
	return r
}

type RemoteCalls struct {
	scope          *types.Scope

	CreationInput  *types.Named
	UpdateInputs   []*types.Named
	DeletionInputs []*types.Named

	ReadOutputs    []*types.Named
}

func (r *RemoteCalls) GetParameterFields() map[string]*types.Var {
	varMap := map[string]*types.Var{}
	c := r.CreationInput.Underlying().(*types.Struct)
	for i := 0; i < c.NumFields(); i++ {
		varMap[c.Field(i).Name()] = c.Field(i)
	}
	for _, upd := range r.UpdateInputs {
		u := upd.Underlying().(*types.Struct)
		for i := 0; i < u.NumFields(); i++ {
			varMap[u.Field(i).Name()] = u.Field(i)
		}
	}
	for _, del := range r.DeletionInputs {
		d := del.Underlying().(*types.Struct)
		for i := 0; i < d.NumFields(); i++ {
			varMap[d.Field(i).Name()] = d.Field(i)
		}
	}
	return varMap
}

func (r *RemoteCalls) GetObservationFields() map[string]*types.Var {
	varMap := map[string]*types.Var{}
	for _, re := range r.ReadOutputs {
		u := re.Underlying().(*types.Struct)
		for i := 0; i < u.NumFields(); i++ {
			varMap[u.Field(i).Name()] = u.Field(i)
		}
	}
	params := r.GetParameterFields()
	for k := range params {
		delete(varMap, k)
	}
	return varMap
}
