package interceptors

import (
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
)

// MountPprof mounts the pprof handlers on the given router at /debug/pprof.
// This should only be used in development or on a protected admin port.
func MountPprof(r chi.Router) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
