package stacktrace

import "fmt"

const (
	DefaultMessageDelimiter = " * "
	DefaultTraceDelimiter   = "\n\n"
	DefaultStackDelimiter   = "\n  |_ "
	DefaultEnsureDuplicates = false
)

type SprintOpt interface {
	Apply(o *SprintOptions)
}

type SprintOptions struct {
	// MessageDelimiter is a delimiter between message and stack trace
	MessageDelimiter string
	// TraceDelimiter is a delimiter between stack traces
	TraceDelimiter string
	// StackDelimiter is a delimiter between stack trace elements
	StackDelimiter string
	// EnsureDuplicates ensures that duplicates are not printed
	EnsureDuplicates bool
	dups             map[string]struct{}
}

func NewSprintOptions() *SprintOptions {
	opts := &SprintOptions{
		EnsureDuplicates: DefaultEnsureDuplicates,
		dups:             make(map[string]struct{}),
		TraceDelimiter:   DefaultTraceDelimiter,
		MessageDelimiter: DefaultMessageDelimiter,
		StackDelimiter:   DefaultStackDelimiter,
	}
	return opts
}

type messageDelimiterOpt string

func (v messageDelimiterOpt) Apply(o *SprintOptions) {
	o.MessageDelimiter = string(v)
}

type traceDelimiterOpt string

func (v traceDelimiterOpt) Apply(o *SprintOptions) {
	o.TraceDelimiter = string(v)
}

type stackDelimiterOpt string

func (v stackDelimiterOpt) Apply(o *SprintOptions) {
	o.StackDelimiter = string(v)
}

type ensureDuplicatesOpt struct{}

func (ensureDuplicatesOpt) Apply(o *SprintOptions) {
	o.EnsureDuplicates = true
}

func WithMessageDelimiter(delimiter string) SprintOpt {
	return messageDelimiterOpt(delimiter)
}

func WithTraceDelimiter(delimiter string) SprintOpt {
	return traceDelimiterOpt(delimiter)
}

func WithStackDelimiter(delimiter string) SprintOpt {
	return stackDelimiterOpt(delimiter)
}

func WithEnsureDuplicates() SprintOpt {
	return &ensureDuplicatesOpt{}
}

func (st *StackTrace) sprint(opts *SprintOptions) string {
	trace := st.Header()
	trace = fmt.Sprintf("%s%s%s", trace, opts.MessageDelimiter, st.FullMessageWithInfo())

	listTraces := ""
	for _, elem := range st.List {
		elemStr := elem.sprint(opts)
		if elemStr == "" {
			continue
		}
		if listTraces != "" {
			listTraces = fmt.Sprintf("%s%s%s", listTraces, opts.TraceDelimiter, elemStr)
		} else {
			listTraces = elemStr
		}
	}

	if _, ok := opts.dups[trace]; ok {
		return listTraces
	}

	if st.Wrapped != nil {
		wrappedStr := st.Wrapped.sprint(opts)
		if wrappedStr == "" {
			return listTraces
		}
		trace = fmt.Sprintf("%s%s%s", trace, opts.StackDelimiter, wrappedStr)
	} else if opts.EnsureDuplicates {
		opts.dups[trace] = struct{}{}
	}

	if listTraces != "" {
		trace = fmt.Sprintf("%s%s%s", trace, opts.TraceDelimiter, listTraces)
	}
	return trace
}

func (st *StackTrace) Sprint(opts ...SprintOpt) string {
	o := NewSprintOptions()
	for _, opt := range opts {
		opt.Apply(o)
	}
	res := st.sprint(o)
	return res
}
