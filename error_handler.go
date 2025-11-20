package gin_error_handler

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorHandler struct {
	errMap          []ErrorMapping
	defaultResponse func(context *gin.Context)
}

// GetMiddleware returns a middleware that handles errors with gin
func (e *ErrorHandler) GetMiddleware() func(c *gin.Context) {
	return func(context *gin.Context) {
		context.Next()
		lastErr := context.Errors.Last()
		if lastErr == nil {
			return
		}
	extFor:
		for _, errorMapping := range e.errMap {
			for _, errToMap := range errorMapping.fromErrors {
				if errors.Is(lastErr.Err, errToMap) ||
					// in this case we cannot use errors.Is, because validator.ValidationErrors does not implement it
					(reflect.TypeOf(errToMap) == reflect.TypeOf(validator.ValidationErrors{}) &&
						reflect.TypeOf(lastErr.Err) == reflect.TypeOf(errToMap)) {
					errorMapping.toResponseFunc(context, lastErr.Err)
					break extFor
				}
			}
		}
		if !context.Writer.Written() {
			e.defaultResponse(context)
		}
	}
}

// NewErrorHandler returns a new ErrorHandler.
// With Options, you can specify the error mappings
// and the default response function. The default response function is mandatory.
func NewErrorHandler(opts Options) (*ErrorHandler, error) {
	err := opts.validate()
	if err != nil {
		return nil, err
	}
	return &ErrorHandler{
		errMap:          opts.errMap,
		defaultResponse: opts.defaultResponse,
	}, nil
}

type ErrorMapping struct {
	fromErrors     []error
	toResponseFunc func(ctx *gin.Context, err error)
}

// ToResponse sets the response function for the error mapping.
func (r ErrorMapping) ToResponse(response func(ctx *gin.Context, err error)) ErrorMapping {
	r.toResponseFunc = response
	return r
}

// Map creates a new ErrorMapping from the given errors.
func Map(err ...error) *ErrorMapping {
	return &ErrorMapping{
		fromErrors: err,
	}
}
