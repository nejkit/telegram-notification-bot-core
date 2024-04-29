package exceptions

import "errors"

var NotFound = errors.New("NotFound")
var OptionalCourseNotSelected = errors.New("OptionalCourseNotSelected")
