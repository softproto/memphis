package common

import (
	"context"
	"errors"
)



var DefFunc = func(ctx context.Context, args interface{}) (interface{}, error) {
	argVal, ok := args.(int)
	if !ok {
		return nil, errors.New("wrong argument type")
	}

	return argVal, nil
}
