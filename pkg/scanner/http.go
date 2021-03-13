package scanner

import "go/types"

func WithIgnoreFns(f ...IgnoreFieldFn) Option {
	return func(rc *RemoteCalls) {
		rc.ignore = f
	}
}

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

type IgnoreFieldFn func(*types.Var) bool

type IgnoreFieldChain []IgnoreFieldFn

func (i IgnoreFieldChain) ShouldIgnore(v *types.Var) bool {
	for _, f := range i {
		if f(v) {
			return true
		}
	}
	return false
}

type RemoteCalls struct {
	scope  *types.Scope
	ignore IgnoreFieldChain // TODO(muvaf): we'll need param/status differentiation.

	CreationInput  *types.Named
	UpdateInputs   []*types.Named
	DeletionInputs []*types.Named

	ReadOutputs []*types.Named
}

func (r *RemoteCalls) AggregatedInput() *types.Struct {
	varMap := map[string]*types.Var{}
	c := r.CreationInput.Underlying().(*types.Struct)
	for i := 0; i < c.NumFields(); i++ {
		if r.ignore.ShouldIgnore(c.Field(i)) {
			continue
		}
		varMap[c.Field(i).Name()] = c.Field(i)
	}
	for _, upd := range r.UpdateInputs {
		u := upd.Underlying().(*types.Struct)
		for i := 0; i < u.NumFields(); i++ {
			if r.ignore.ShouldIgnore(u.Field(i)) {
				continue
			}
			varMap[u.Field(i).Name()] = u.Field(i)
		}
	}
	for _, del := range r.DeletionInputs {
		d := del.Underlying().(*types.Struct)
		for i := 0; i < d.NumFields(); i++ {
			if r.ignore.ShouldIgnore(d.Field(i)) {
				continue
			}
			varMap[d.Field(i).Name()] = d.Field(i)
		}
	}
	fields := make([]*types.Var, len(varMap))
	i := 0
	for _, v := range varMap {
		fields[i] = v
		i++
	}
	return types.NewStruct(fields, nil)
}

func (r *RemoteCalls) AggregatedOutput() *types.Struct {
	varMap := map[string]*types.Var{}
	for _, re := range r.ReadOutputs {
		ro := re.Underlying().(*types.Struct)
		for i := 0; i < ro.NumFields(); i++ {
			if r.ignore.ShouldIgnore(ro.Field(i)) {
				continue
			}
			varMap[ro.Field(i).Name()] = ro.Field(i)
		}
	}
	params := r.AggregatedInput()
	for i := 0; i < params.NumFields(); i++ {
		delete(varMap, params.Field(i).Name())
	}
	fields := make([]*types.Var, len(varMap))
	i := 0
	for _, v := range varMap {
		fields[i] = v
		i++
	}
	return types.NewStruct(fields, nil)
}
