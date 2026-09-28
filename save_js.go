//go:build js

package collider

import "syscall/js"

// In the browser there is no filesystem: saves go to localStorage,
// under "<game title>/<key>" so games sharing an origin (itch.io, a
// local dev server) keep separate data.

// localStorage returns the storage object, or false where the page may
// not use it (a sandboxed iframe or blocked site data throws on access).
func localStorage() (st js.Value, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	st = js.Global().Get("localStorage")
	return st, st.Truthy()
}

func writeSave(title, key string, b []byte) (err error) {
	st, ok := localStorage()
	if !ok {
		return errNoStorage
	}
	defer func() {
		if recover() != nil { // quota exceeded, storage disabled
			err = errNoStorage
		}
	}()
	st.Call("setItem", title+"/"+key, string(b))
	return nil
}

func readSave(title, key string) (b []byte, ok bool) {
	st, ok := localStorage()
	if !ok {
		return nil, false
	}
	defer func() {
		if recover() != nil {
			b, ok = nil, false
		}
	}()
	v := st.Call("getItem", title+"/"+key)
	if v.IsNull() || v.IsUndefined() {
		return nil, false
	}
	return []byte(v.String()), true
}
