package http

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/protobuf/proto"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/pkg/logger"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

func (h *Server) generateJWT(userID int64, username string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.config.JwtKey)
}

func (h *Server) validateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return h.config.JwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.UserID == 0 || len(claims.Username) == 0 {
		return nil, errors.New("invalid userID or username")
	}
	return claims, nil
}

type SignupRequest struct {
	Username    string `json:"user_name"`
	Password    string `json:"password"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Dob         string `json:"dob"`
}

type UserData struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Dob         string `json:"dob"`
}

type UserDataWithToken struct {
	UserData
	Token string `json:"token"`
}

func (h *Server) Signup(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req = &SignupRequest{}
	)

	// bind req
	if err := c.ShouldBind(req); err != nil {
		bindErr := common.WrapError(common.CodeInvalidRequest, "bind request error", err)
		h.returnErrResp(c, bindErr)
		return
	}

	// validate req
	if err := validateSignupReq(req); err != nil {
		validateErr := common.WrapError(common.CodeInvalidRequest, "invalid request", err)
		h.returnErrResp(c, validateErr)
		return
	}

	// process logic
	grpcReq := &grpc.SignupRequest{
		UserName:    proto.String(req.Username),
		Password:    proto.String(req.Password),
		DisplayName: proto.String(req.DisplayName),
		Email:       proto.String(req.Email),
		Dob:         proto.String(req.Dob),
	}

	grpcResp, err := h.grpcClient.Signup(ctx, grpcReq)
	if err != nil {
		appErr := common.FromGRPCError(err)
		h.returnErrResp(c, appErr)
		return
	}

	// process response
	createdUser := grpcResp.GetUser()
	userData := &UserData{
		ID:          createdUser.GetId(),
		Username:    createdUser.GetUserName(),
		Email:       createdUser.GetEmail(),
		DisplayName: createdUser.GetDisplayName(),
		Dob:         createdUser.GetDob(),
	}

	h.returnDataResp(c, "Sign up successfully", userData)
}

func validateSignupReq(req *SignupRequest) error {
	if len(req.Username) <= 4 {
		return fmt.Errorf("username must have at least 4 characters")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must have at least 8 characters")
	}
	if match, _ := regexp.Match(`^.+@.+\..+$`, []byte(req.Email)); !match { // simple check: xxx@xxx.xxx
		return fmt.Errorf("invalid email format")
	}
	if len(req.DisplayName) <= 4 {
		return fmt.Errorf("display_name must have at least 4 characters")
	}

	dobTime, err := time.Parse("20060102", req.Dob)
	if err != nil {
		return fmt.Errorf("dob must be a valid date with YYYYMMDD format")
	}
	if dobTime.Year() < 1900 {
		return fmt.Errorf("invalid DOB year")
	}

	return nil
}

type LoginRequest struct {
	Username string `json:"user_name"`
	Password string `json:"password"`
}

func (h *Server) Login(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req = &LoginRequest{}
		api = c.Request.Method + " " + c.Request.RequestURI
	)

	// bind req
	if err := c.ShouldBind(req); err != nil {
		bindErr := common.WrapError(common.CodeInvalidRequest, "bind request error", err)
		h.returnErrResp(c, bindErr)
		return
	}
	logger.Debug("parse request", logger.F("api", api), logger.F("req", req))

	// validate req etc ...

	// process logic
	grpcReq := &grpc.LoginRequest{
		UserName: proto.String(req.Username),
		Password: proto.String(req.Password),
	}

	grpcResp, err := h.grpcClient.Login(ctx, grpcReq)
	if err != nil {
		appErr := common.FromGRPCError(err)
		h.returnErrResp(c, appErr)
		return
	}

	// process response
	createdUser := grpcResp.GetUser()
	token, err := h.generateJWT(createdUser.GetId(), createdUser.GetUserName(), 24*time.Hour)
	if err != nil {
		tokenErr := common.WrapError(common.CodeInternal, "generate jwt token error", err)
		h.returnErrResp(c, tokenErr)
		return
	}

	userData := &UserDataWithToken{
		UserData: UserData{
			ID:          createdUser.GetId(),
			Username:    createdUser.GetUserName(),
			Email:       createdUser.GetEmail(),
			DisplayName: createdUser.GetDisplayName(),
			Dob:         createdUser.GetDob(),
		},
		Token: token,
	}

	h.returnDataResp(c, "Log in successfully", userData)
}
