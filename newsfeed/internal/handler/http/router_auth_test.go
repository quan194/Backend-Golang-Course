package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/internal/handler/proto/grpc"
)

func TestServer_Signup(t *testing.T) {
	// assume
	mockUserClient := new(grpc.MockServiceClient)
	mockUserClient.On("Signup", mock.Anything, mock.AnythingOfType("*grpc.SignupRequest")).
		Return(&grpc.SignupResponse{
			User: &grpc.UserData{
				Id:          proto.Int64(1),
				UserName:    proto.String("username"),
				Email:       proto.String("abc@gmai.com"),
				DisplayName: proto.String("displayname"),
				Dob:         proto.String("19990101"),
			},
		}, nil)

	cfg := Config{
		Host: "127.0.0.1",
		Port: 18080,
	}
	srv, err := New(cfg, mockUserClient)
	assert.NoError(t, err)

	// act
	body := `{
		"user_name": "username",
		"password": "password",
		"email": "abc@gmai.com",
		"display_name": "displayname",
		"dob": "19990101"
	}`

	req := httptest.NewRequest(http.MethodPost, "/grpc/signup", bytes.NewBuffer([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	// assert
	rec := httptest.NewRecorder()
	srv.router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	resp := new(DataResponse)
	err = json.Unmarshal(rec.Body.Bytes(), resp)
	assert.NoError(t, err)
	assert.Equal(t, common.CodeOK, resp.Code)

	// TODO: because we use resp.Data as interface{} type, so after json marshal, its default type is map[string]interface{}
	// Can consider to use specific struct instead of interface{} in resp, so it is easier to parse to struct here
	userDataJson, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, int(userDataJson["id"].(float64)))
	assert.Equal(t, "username", userDataJson["username"])
}
