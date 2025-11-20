package gin_error_handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNewErrorHandler(t *testing.T) {
	t.Run("should return a new ErrorHandler with default response function", func(t *testing.T) {
		options := &Options{}
		options.DefaultResponse(func(context *gin.Context) {})
		eh, err := NewErrorHandler(*options)
		assert.NoError(t, err)
		assert.NotNil(t, eh)
		assert.NotNil(t, eh.defaultResponse)
	})
	t.Run("should return an error", func(t *testing.T) {
		options := &Options{}
		eh, err := NewErrorHandler(*options)
		assert.Error(t, err)
		assert.Nil(t, eh)
	})
	t.Run("should return a new ErrorHandler with error mappings", func(t *testing.T) {
		options := &Options{}
		options.DefaultResponse(func(context *gin.Context) {})
		options.ErrorMappings([]ErrorMapping{
			Map(assert.AnError).ToResponse(func(ctx *gin.Context, err error) {}),
		})
		eh, err := NewErrorHandler(*options)
		assert.NoError(t, err)
		assert.NotNil(t, eh)
		assert.Len(t, eh.errMap, 1)
	})

}

type customError struct{}

func (e customError) Error() string {
	return "custom error"
}
func (e customError) Is(target error) bool {
	var customError customError
	ok := errors.As(target, &customError)
	return ok
}

func TestErrorHandler_GetMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should call mapped response when error matches", func(t *testing.T) {
		testErr := errors.New("test error")
		responseCalled := false

		options := Options{}
		options.DefaultResponse(func(context *gin.Context) {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "default"})
		})
		options.ErrorMappings([]ErrorMapping{
			Map(testErr).ToResponse(func(ctx *gin.Context, err error) {
				responseCalled = true
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}),
		})

		eh, err := NewErrorHandler(options)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		_ = c.Error(testErr)

		middleware := eh.GetMiddleware()
		middleware(c)

		assert.True(t, responseCalled)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"test error"}`, w.Body.String())
	})

	t.Run("should call default response when no error matches", func(t *testing.T) {
		testErr := errors.New("test error")
		unmappedErr := errors.New("unmapped error")

		options := &Options{}
		options.DefaultResponse(func(context *gin.Context) {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "default"})
		})
		options.ErrorMappings([]ErrorMapping{
			Map(testErr).ToResponse(func(ctx *gin.Context, err error) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}),
		})

		eh, err := NewErrorHandler(*options)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		_ = c.Error(unmappedErr)

		middleware := eh.GetMiddleware()
		middleware(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, `{"error":"default"}`, w.Body.String())
	})

	t.Run("should not handle when no error present", func(t *testing.T) {
		options := &Options{}
		options.DefaultResponse(func(context *gin.Context) {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "default"})
		})

		eh, err := NewErrorHandler(*options)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		middleware := eh.GetMiddleware()
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.False(t, c.Writer.Written())
	})

	t.Run("should handle wrapped errors with errors.Is", func(t *testing.T) {
		baseErr := errors.New("base error")
		wrappedErr := errors.Join(baseErr, errors.New("additional context"))

		options := &Options{}
		options.DefaultResponse(func(context *gin.Context) {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "default"})
		})
		options.ErrorMappings([]ErrorMapping{
			Map(baseErr).ToResponse(func(ctx *gin.Context, err error) {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "mapped"})
			}),
		})

		eh, err := NewErrorHandler(*options)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		_ = c.Error(wrappedErr)

		middleware := eh.GetMiddleware()
		middleware(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"mapped"}`, w.Body.String())
	})

	t.Run("should use first matching error mapping", func(t *testing.T) {
		testErr := errors.New("test error")
		firstCalled := false
		secondCalled := false

		options := &Options{}
		options.DefaultResponse(func(context *gin.Context) {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "default"})
		})
		options.ErrorMappings([]ErrorMapping{
			Map(testErr).ToResponse(func(ctx *gin.Context, err error) {
				firstCalled = true
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "first"})
			}),
			Map(testErr).ToResponse(func(ctx *gin.Context, err error) {
				secondCalled = true
				ctx.JSON(http.StatusConflict, gin.H{"error": "second"})
			}),
		})

		eh, err := NewErrorHandler(*options)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		_ = c.Error(testErr)

		middleware := eh.GetMiddleware()
		middleware(c)

		assert.True(t, firstCalled)
		assert.False(t, secondCalled)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"first"}`, w.Body.String())
	})
	t.Run("should handle custom error", func(t *testing.T) {
		testErr := customError{}
		responseCalled := false

		options := Options{}
		options.DefaultResponse(func(context *gin.Context) {
			context.JSON(http.StatusInternalServerError, gin.H{"error": "default"})
		})
		options.ErrorMappings([]ErrorMapping{
			Map(customError{}).ToResponse(func(ctx *gin.Context, err error) {
				responseCalled = true
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			}),
		})

		eh, err := NewErrorHandler(options)
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		_ = c.Error(testErr)

		middleware := eh.GetMiddleware()
		middleware(c)

		assert.True(t, responseCalled)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, `{"error":"custom error"}`, w.Body.String())
	})
}
