package utils

import (
	"context"
	"log"
	"runtime"
)

func HandlePanic(ctx context.Context, f func(...interface{})) func(...interface{}) {
	return func(arg ...interface{}) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 64<<10)
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				log.Printf("%v: %+v\n%s\n", err, arg, buf)
			}
		}()

		f(arg...)
	}
}
