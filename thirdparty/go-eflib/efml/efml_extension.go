/////////////////////////////////////////////////////////////////////////////
//  @brief API interface of Enflame Managerment Library
//
//  Enflame Tech, All Rights Reserved. 2022 Copyright (C)
/////////////////////////////////////////////////////////////////////////////

package efml

// #cgo LDFLAGS: -ldl  -Wl,--unresolved-symbols=ignore-in-object-files
// #include "stdbool.h"
// #include "efml.h"
import "C"

/*
 * @brief Enfalme Management Library get the device low power mode.
 */
func (h Handle) GetDevIsLowPowerMode() (bool, error) {
	var is_low_power_mode C.bool
	r := C.EfmlGetDevIsLowPowerMode(C.uint(h.Dev_Idx), &is_low_power_mode)
	if r == C.EFML_ERROR_NOT_SUPPORTED {
		return false, nil
	}

	return bool(is_low_power_mode), errorString(r)
}
