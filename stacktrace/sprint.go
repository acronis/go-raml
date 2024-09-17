package stacktrace

import "fmt"

func (st *StackTrace) Sprint() string {
	str := st.Header()
	str = fmt.Sprintf("%s\n\t%s", str, st.FullMessage())
	if st.Wrapped != nil {
		str = fmt.Sprintf("%s\n%s", str, st.Wrapped.Sprint())
	}
	return str
}
