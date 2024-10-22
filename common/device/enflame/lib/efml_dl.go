/////////////////////////////////////////////////////////////////////////////
//  @brief API interface of Enflame Managerment Library
//
//  Enflame Tech, All Rights Reserved. 2019 Copyright (C)
/////////////////////////////////////////////////////////////////////////////

package lib

import (
	"unsafe"
)

/*
#include <dlfcn.h>
#include <stdbool.h>
#include <string.h>
#include "efml.h"
*/
import "C"

type dl_handle_ptr struct{ handles []unsafe.Pointer }

var dl dl_handle_ptr

func (dl *dl_handle_ptr) Init() C.efmlReturn_t {
	handle := C.dlopen(C.CString("libefml.so"), C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == C.NULL {
		handle := C.dlopen(C.CString("libefml.so.1.0.0"), C.RTLD_LAZY|C.RTLD_GLOBAL)
		if handle == C.NULL {
			handle := C.dlopen(C.CString("/usr/lib/libefml.so"), C.RTLD_LAZY|C.RTLD_GLOBAL)
			if handle == C.NULL {
				handle := C.dlopen(C.CString("/usr/lib/libefml.so.1.0.0"), C.RTLD_LAZY|C.RTLD_GLOBAL)
				if handle == C.NULL {
					return C.EFML_ERROR_LIBRARY_NOT_FOUND
				}
			}
		}
	}

	dl.handles = append(dl.handles, handle)
	no_driver := true // do not use kmd
	return C.EfmlInit(C.bool(no_driver))
}

func (dl *dl_handle_ptr) InitV2(no_driver bool) C.efmlReturn_t {
	handle := C.dlopen(C.CString("libefml.so"), C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == C.NULL {
		handle := C.dlopen(C.CString("/usr/lib/libefml.so"), C.RTLD_LAZY|C.RTLD_GLOBAL)
		if handle == C.NULL {
			handle := C.dlopen(C.CString("/usr/lib/libefml.so.1.0.0"), C.RTLD_LAZY|C.RTLD_GLOBAL)
			if handle == C.NULL {
				return C.EFML_ERROR_LIBRARY_NOT_FOUND
			}
		}
	}

	dl.handles = append(dl.handles, handle)

	return C.EfmlInit(C.bool(no_driver))
}

func (dl *dl_handle_ptr) Shutdown() C.efmlReturn_t {
	C.EfmlShutdown()

	for _, handle := range dl.handles {
		err := C.dlclose(handle)
		if err != 0 {
			return C.EFML_ERROR_FUNCTION_NOT_FOUND
		}
	}

	return C.EFML_SUCCESS
}

func (dl *dl_handle_ptr) lookupSymbol(symbol string) C.efmlReturn_t {
	for _, handle := range dl.handles {
		C.dlerror()
		C.dlsym(handle, C.CString(symbol))
		if unsafe.Pointer(C.dlerror()) == C.NULL {
			return C.EFML_SUCCESS
		}
	}

	return C.EFML_ERROR_FUNCTION_NOT_FOUND
}
