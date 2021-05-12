package db

import (
	"fmt"
)

var ErrNotFound = fmt.Errorf("not found")
var ErrRecordExists = fmt.Errorf("record exists")
var ErrNotAuthorized = fmt.Errorf("not authorized")
