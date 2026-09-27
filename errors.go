package mwanachamaworkspace

import "errors"

var ErrInvalidWorkbook = errors.New("workspace: a workbook needs an owner and a title")

var ErrWorkbookNotFound = errors.New("workspace: no such workbook")

var ErrInvalidReference = errors.New("workspace: a reference needs a workbook, a module and a label")

var ErrReferenceNotFound = errors.New("workspace: no such reference")

var ErrInvalidUse = errors.New("workspace: a use needs a workbook and a well-formed action")

var ErrInvalidAction = errors.New("workspace: that is not a well-formed action")

var ErrInvalidShare = errors.New("workspace: a share needs a workbook and a key")

var ErrShareNotFound = errors.New("workspace: no such share")

var ErrShareRetired = errors.New("workspace: that key no longer opens this workbook")
