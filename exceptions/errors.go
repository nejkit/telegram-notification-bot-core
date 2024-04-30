package exceptions

import "errors"

var NotFound = errors.New("NotFound")
var InternalError = errors.New("InternalError")
var AlreadyExists = errors.New("AlreadyExists")
var OptionalCourseNotSelected = errors.New("O	ptionalCourseNotSelected")
