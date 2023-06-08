package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/lib/pq"
	mockdb "github.com/patchbrain/simple-bank/db/mock"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	if err := util.CheckPassword(e.password, arg.PasswordHashed); err != nil {
		return false
	}
	e.arg.PasswordHashed = arg.PasswordHashed
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("is equal to %v", e.arg)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{
		arg:      arg,
		password: password,
	}
}

func randomUser(t *testing.T) (db.User, string) {
	password := util.RandomString(6)
	hashed, err := util.HashPassword(password)
	require.NoError(t, err)
	user := db.User{
		PasswordHashed: hashed,
		Username:       util.RandomString(6),
		FullName:       util.RandomString(3),
		Email:          util.RandomEmail(),
	}
	return user, password
}

func TestLoginUser(t *testing.T) {
	user, password := randomUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		createStub    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Eq(user.Username)).Times(1).Return(user, nil)
				store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UserNotFound",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username": user.Username,
				"password": password,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetUser(gomock.Any(), gomock.Any()).Times(1).Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// 建立stub
			// 代表GetAccount函数必须执行一次，且执行所返回的值是account，err
			tc.createStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder() // http响应记录器

			// 应该符合api的uri
			url := "/users/login"
			bodyBytes, err := json.Marshal(createUserRequest{
				Username: tc.body["username"].(string),
				Password: tc.body["password"].(string),
			})
			require.NoError(t, err)

			body := bytes.NewBuffer(bodyBytes)

			request, err := http.NewRequest(http.MethodPost, url, body)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, request) // 处理请求，stub起作用
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateUser(t *testing.T) {
	// 一组正确的参数
	user, password := randomUser(t)
	passwordShort := password[:3]

	testCases := []struct {
		name          string
		body          gin.H
		createStub    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       user.Username,
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          user.Email,
				}, password)).Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "ErrorEmail",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     "123",
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       user.Username,
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          "123",
				}, password)).Times(0).Return(db.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ErrorUsername",
			body: gin.H{
				"username":  "user.Username",
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       "user.Username",
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          user.Email,
				}, password)).Times(0).Return(db.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFound",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       user.Username,
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          user.Email,
				}, password)).Times(1).Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       user.Username,
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          user.Email,
				}, password)).Times(1).Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "ShortPassword",
			body: gin.H{
				"username":  user.Username,
				"password":  passwordShort,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       user.Username,
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          user.Email,
				}, password)).Times(0).Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "Duplicate",
			body: gin.H{
				"username":  user.Username,
				"password":  password,
				"full_name": user.FullName,
				"email":     user.Email,
			},
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), EqCreateUserParams(db.CreateUserParams{
					Username:       user.Username,
					PasswordHashed: user.PasswordHashed,
					FullName:       user.FullName,
					Email:          user.Email,
				}, password)).Times(1).Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			// 建立stub
			// 代表GetAccount函数必须执行一次，且执行所返回的值是account，err
			tc.createStub(store)

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder() // http响应记录器

			// 应该符合api的uri
			url := "/users"
			bodyBytes, err := json.Marshal(createUserRequest{
				Username: tc.body["username"].(string),
				Password: tc.body["password"].(string),
				FullName: tc.body["full_name"].(string),
				Email:    tc.body["email"].(string),
			})
			require.NoError(t, err)

			body := bytes.NewBuffer(bodyBytes)

			request, err := http.NewRequest(http.MethodPost, url, body)
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, request) // 处理请求，stub起作用
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var userRsp = UserResponse{}
	err = json.Unmarshal(data, &userRsp)
	require.NoError(t, err)

	require.Equal(t, user.Username, userRsp.Username)
	require.Equal(t, user.Email, userRsp.Email)
	require.Equal(t, user.FullName, userRsp.FullName)
	require.WithinDuration(t, user.CreatedAt, userRsp.CreatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, userRsp.PasswordChangedAt, time.Second)
}
