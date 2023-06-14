package util

import "context"

type key int

const (
	RecursionCountContextKey key = 1000
)

func GetContextRecursionCount(ctx context.Context) int {
	currentValue, ok := ctx.Value(RecursionCountContextKey).(int)
	if !ok {
		return 0
	}
	return currentValue
}

func SetContextRecursionCount(ctx context.Context, value int) context.Context {
	return context.WithValue(ctx, RecursionCountContextKey, value)
}

func IncrementContextRecursionCount(ctx context.Context) context.Context {
	currentValue := GetContextRecursionCount(ctx)
	currentValue++
	return SetContextRecursionCount(ctx, currentValue)
}
