package scanner

import "go/types"

func WithReadCalls(s []string) Option {
	return func(r *RemoteCalls) {
		for _, typeName := range s {
			if t := r.scope.Lookup(typeName); t != nil {
				r.Reads = append(r.Reads, t.Type().(*types.Named))
			}
		}
	}
}

func WithCreateCall(s string) Option {
	return func(r *RemoteCalls) {
		if t := r.scope.Lookup(s); t != nil {
			r.Creation = t.Type().(*types.Named)
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
	scope    *types.Scope
	Reads    []*types.Named
	Creation *types.Named
	Updates  []*types.Named
	Deletes  []*types.Named
}

func (r *RemoteCalls) GetParameterFields() map[string]*types.Var {
	varMap := map[string]*types.Var{}
	c := r.Creation.Underlying().(*types.Struct)
	for i := 0; i < c.NumFields(); i++ {
		varMap[c.Field(i).Name()] = c.Field(i)
	}
	for _, upd := range r.Updates {
		u := upd.Underlying().(*types.Struct)
		for i := 0; i < u.NumFields(); i++ {
			varMap[u.Field(i).Name()] = u.Field(i)
		}
	}
	for _, del := range r.Deletes {
		d := del.Underlying().(*types.Struct)
		for i := 0; i < d.NumFields(); i++ {
			varMap[d.Field(i).Name()] = d.Field(i)
		}
	}
	return varMap
}

func (r *RemoteCalls) GetObservationFields() map[string]*types.Var {
	varMap := map[string]*types.Var{}
	for _, re := range r.Reads {
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
