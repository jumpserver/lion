package common

import (
	"github.com/gofrs/uuid"
)

/*
https://github.com/satori/go.uuid/issues/84
替换 uuid 包为 github.com/gofrs/uuid
*/

func UUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}
