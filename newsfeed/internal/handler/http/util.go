package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/pkg/logger"
)

var errCodeToHttpCodeMap = map[common.ErrorCode]int{
	// 0: ok
	common.CodeOK: http.StatusOK,
	// 1xx: client error
	common.CodeInvalidRequest: http.StatusBadRequest,
	common.CodeUnauthorized:   http.StatusUnauthorized,
	// 2xx: biz err
	common.CodeInvalidLogin:       http.StatusBadRequest,
	common.CodeExistedUsername:    http.StatusBadRequest,
	common.CodeNotExistedUsername: http.StatusBadRequest,
	common.CodeNotExistedUserID:   http.StatusBadRequest,

	// 9xx: internal error
	common.CodeInternal:       http.StatusInternalServerError,
	common.CodeNotImplemented: http.StatusInternalServerError,
}

type ErrorResponse struct {
	Code    common.ErrorCode `json:"code"`
	Message string           `json:"message"`
}

type DataResponse struct {
	Code    common.ErrorCode `json:"code"`
	Message string           `json:"message"`
	Data    interface{}      `json:"data"`
}

func getHttpStatus(err *common.AppError) int {
	httpStatus, ok := errCodeToHttpCodeMap[err.Code]
	if !ok {
		return http.StatusInternalServerError
	}
	return httpStatus
}

func getErrMsg(err *common.AppError) string {
	msg := err.Message
	if err.Code >= common.CodeInternal { // for internal error, hide it from users
		msg = "internal server error"
	}
	return msg
}

func (h *Server) returnErrResp(c *gin.Context, err error) {
	api := c.FullPath()
	appError, ok := err.(*common.AppError)
	if !ok {
		appError = common.WrapError(common.CodeInternal, "unknown error", err)
	}

	httpStatus := getHttpStatus(appError)
	errMsg := getErrMsg(appError)

	c.JSON(httpStatus, &ErrorResponse{
		Code:    appError.Code,
		Message: errMsg,
	})

	logger.Error("err response",
		logger.F("api", api),
		logger.F("http_code", httpStatus),
		logger.F("resp_code", appError.Code),
		logger.E(err),
	)
}

func (h *Server) returnDataResp(c *gin.Context, msg string, data any) {
	var (
		api = c.Request.Method + " " + c.Request.RequestURI
	)

	c.JSON(http.StatusOK, &DataResponse{
		Code:    common.CodeOK,
		Message: msg,
		Data:    data,
	})

	logger.Info("data response",
		logger.F("api", api),
		logger.F("http_code", http.StatusOK),
		logger.F("resp_code", common.CodeOK),
		logger.F("data", data),
	)
}
