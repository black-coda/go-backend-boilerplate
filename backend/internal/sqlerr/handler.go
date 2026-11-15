package sqlerr

import "errors"

func ErrCode(err error) Code{
	var pgerr *Error

	if errors.As(err, &pgerr){
		return pgerr.Code
	}
	return Other
}