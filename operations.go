package mwanachamaworkspace

import _ "embed"

//go:embed workspace.operations.json
var operationsJSON []byte

func Operations() []byte { return operationsJSON }
