package gin_error_handler

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Options struct {
	errMap          []ErrorMapping
	defaultResponse func(context *gin.Context)
}

func (o *Options) ErrorMappings(m []ErrorMapping) *Options {
	o.errMap = m
	return o
}

func (o *Options) DefaultResponse(f func(context *gin.Context)) *Options {
	o.defaultResponse = f
	return o
}

func (o *Options) validate() error {
	if o.defaultResponse == nil {
		return errors.New("defaultResponse is required")
	}
	return nil
}
