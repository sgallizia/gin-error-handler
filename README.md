# gin-error-handler

A middleware to handle errors in Gin web framework by mapping them to HTTP responses.

## Quick Start

### Installation

```bash
go get github.com/sgallizia/gin-error-handler
```

### Example

```go
package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	ginErrorHandler "github.com/sgallizia/gin-error-handler"
)

var stdError = errors.New("error")

type customError struct {
}

func (e customError) Error() string {
	return "custom error"
}

func (e customError) Is(target error) bool {
	var cErr *customError
	ok := errors.As(target, &cErr)
	return ok
}

func main() {
	engine := gin.Default()
	opts := ginErrorHandler.Options{}
	opts.DefaultResponse(func(context *gin.Context) {
		context.JSON(500, gin.H{"error": "internal server error"})
	})
	opts.ErrorMappings([]ginErrorHandler.ErrorMapping{
		ginErrorHandler.Map(stdError).ToResponse(func(context *gin.Context, err error) {
			context.JSON(http.StatusBadRequest, gin.H{"error": "standard error occurred"})
		}),
		ginErrorHandler.Map(customError{}).ToResponse(func(context *gin.Context, err error) {
			context.JSON(http.StatusTeapot, gin.H{"error": "custom error occurred"})
		}),
	})
	errorHandlerMdl, err := ginErrorHandler.NewErrorHandler(opts)
	if err != nil {
		panic(err)
	}
	engine.Use(errorHandlerMdl.GetMiddleware())
	engine.GET("/ping", func(c *gin.Context) {
		_ = c.Error(customError{})
	})
	engine.GET("/pong", func(c *gin.Context) {
		_ = c.Error(stdError)
	})
	err = engine.Run()
	if err != nil {
		panic(err)
	}
}
```
