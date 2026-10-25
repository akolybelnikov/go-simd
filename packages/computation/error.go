package computation

import "errors"

var BadInputError = errors.New("input slices must have the same length and match the output slice length")
