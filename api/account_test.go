package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	mockdb "github.com/patchbrain/simple-bank/db/mock"
	db "github.com/patchbrain/simple-bank/db/sqlc"
	"github.com/patchbrain/simple-bank/token"
	"github.com/patchbrain/simple-bank/util"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func randomAccount(username string) db.Account {
	return db.Account{
		ID:       util.RandomInt64(1, 1000),
		Owner:    username,
		Balance:  util.RandomInt64(1, 2000),
		Currency: util.RandomCurrency(),
	}
}

func TestGetAccount(t *testing.T) {
	user, _ := randomUser()
	account := randomAccount(user.Username)
	testCases := []struct {
		name            string
		accountId       int64
		createStub      func(store *mockdb.MockStore)
		setupAuthHeader func(t *testing.T, r *http.Request, maker token.Maker)
		checkResponse   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountId: account.ID,
			createStub: func(store *mockdb.MockStore) {
				// 从连接池中获取了一个连接，但在使用他之前就关闭了连接，当使用该连接时就会报错ErrConnDone
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			// 给了一个错误的accountID
			name:      "BadRequest",
			accountId: 0,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(0)).Times(0)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "No AuthHeader",
			accountId: 0,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(0)).Times(0)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuthHeader(t, request, server.TokenMaker)

			server.Router.ServeHTTP(recorder, request) // 处理请求，stub起作用
			tc.checkResponse(t, recorder)
		})
	}
}

func TestListAccount(t *testing.T) {
	// 创建一组accounts
	accounts := make([]db.Account, 0)
	user, _ := randomUser()
	n := 5
	for i := 0; i < n; i++ {
		account := randomAccount(user.Username)
		accounts = append(accounts, account)
	}
	testCases := []struct {
		name            string
		pageId          int32
		pageSize        int32
		createStub      func(store *mockdb.MockStore)
		setupAuthHeader func(t *testing.T, r *http.Request, maker token.Maker)
		checkResponse   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			pageId:   1,
			pageSize: 5,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccount(gomock.Any(), gomock.Eq(db.ListAccountParams{
					Owner:  user.Username,
					Limit:  5,
					Offset: 0,
				})).Times(1).Return(accounts[:5], nil)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccountList(t, recorder.Body, accounts[:5])
			},
		},
		{
			name:     "NotFound",
			pageId:   3,
			pageSize: 5,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccount(gomock.Any(), gomock.Eq(db.ListAccountParams{
					Owner:  user.Username,
					Limit:  5,
					Offset: 10,
				})).Times(1).Return([]db.Account{}, sql.ErrNoRows)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:     "InternalError",
			pageId:   1,
			pageSize: 5,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccount(gomock.Any(), gomock.Eq(db.ListAccountParams{
					Owner:  user.Username,
					Limit:  5,
					Offset: 0,
				})).Times(1).Return([]db.Account{}, sql.ErrConnDone)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:     "BadRequest",
			pageId:   0,
			pageSize: 5,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, user.Username, time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "No AuthHeader",
			pageId:   1,
			pageSize: 5,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
			u, err := url.ParseRequestURI("/accounts")
			require.NoError(t, err)
			q := u.Query()
			q.Set("page_id", strconv.Itoa(int(tc.pageId)))
			q.Set("page_size", strconv.Itoa(int(tc.pageSize)))
			u.RawQuery = q.Encode()
			//url := fmt.Sprintf("/accounts?pageId=%d&pageSize=%d", tc.pageId, tc.pageSize)
			request, err := http.NewRequest(http.MethodGet, u.String(), nil)
			require.NoError(t, err)

			tc.setupAuthHeader(t, request, server.TokenMaker)

			server.Router.ServeHTTP(recorder, request) // 处理请求，stub起作用
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccount(t *testing.T) {
	account := db.Account{
		ID:       1,
		Owner:    "Hno3",
		Balance:  0,
		Currency: "USD",
	}

	testCases := []struct {
		name            string
		owner           string
		currency        string
		createStub      func(store *mockdb.MockStore)
		setupAuthHeader func(t *testing.T, r *http.Request, maker token.Maker)
		checkResponse   func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			owner:    account.Owner,
			currency: account.Currency,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				})).Times(1).Return(account, nil)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, "Hno3", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, account)
			},
		}, {
			name:     "InternalError",
			owner:    account.Owner,
			currency: account.Currency,
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
					Owner:    account.Owner,
					Balance:  0,
					Currency: account.Currency,
				})).Times(1).Return(account, sql.ErrConnDone)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, "Hno3", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		}, {
			name:     "BadRequest",
			owner:    account.Owner,
			currency: "",
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
				addAuthHeader(t, maker, r, authTypeBearer, "Hno3", time.Minute)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		}, {
			name:     "No AuthHeader",
			owner:    account.Owner,
			currency: "",
			createStub: func(store *mockdb.MockStore) {
				store.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Times(0)
			},
			setupAuthHeader: func(t *testing.T, r *http.Request, maker token.Maker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
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
			url := "/accounts"
			bodyBytes, err := json.Marshal(createAccountRequest{
				Owner:    tc.owner,
				Currency: tc.currency,
			})
			require.NoError(t, err)

			body := bytes.NewBuffer(bodyBytes)

			request, err := http.NewRequest(http.MethodPost, url, body)
			require.NoError(t, err)

			tc.setupAuthHeader(t, request, server.TokenMaker)

			server.Router.ServeHTTP(recorder, request) // 处理请求，stub起作用
			tc.checkResponse(t, recorder)
		})
	}
}

// requireBodyMatchAccount 用于验证响应体的内容是否与targetAccount对应，若对应则返回true，不对应返回false
func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, targetAccount db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, gotAccount, targetAccount)
}

// requireBodyMatchAccountList 用于验证响应体的内容是否与account切片对应，若对应则返回true，不对应返回false
func requireBodyMatchAccountList(t *testing.T, body *bytes.Buffer, targetAccounts []db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount []db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, gotAccount, targetAccounts)
}
