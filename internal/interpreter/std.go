package interpreter

import (
	"context"
	"time"
)

func StdFnTime(ctx context.Context, interpeter *interpreter) (any, error) {
	return time.Now().UnixMilli(), nil
}

func StdFnPPrint(ctx context.Context, interpeter *interpreter, args ...any) (any, error) {
	interpeter.print(args...)
	return nil, nil
}
