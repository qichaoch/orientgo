package obinary

import (
	"github.com/qichaoch/orientgo/obinary/rw"
)

func ReadErrorResponse(r *rw.Reader) (serverException error) {
	return readErrorResponse(r, CurrentProtoVersion)
}
