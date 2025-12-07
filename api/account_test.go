package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	mockdb "github.com/roman-adamchik/simplebank/db/mock"
	db "github.com/roman-adamchik/simplebank/db/sqlc"
	"github.com/roman-adamchik/simplebank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountAPI(t *testing.T) {
	account := getRandomAccount()

	testCases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, &account)
			},
		},
		{
			name:      "NotFound",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "BadRequestFailValidation",
			accountId: 0,
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {

				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			request, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestCreateAccountAPI(t *testing.T) {
	account := getRandomAccount()

	testCases := []struct {
		name          string
		owner         string
		currency      string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			owner:    account.Owner,
			currency: account.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						Owner:    account.Owner,
						Balance:  0,
						Currency: account.Currency,
					})).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccount(t, recorder.Body, &account)
			},
		},
		{
			name:     "BadRequestOwnerMissing",
			owner:    "",
			currency: account.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "BadRequestCurrencyInvalid",
			owner:    account.Owner,
			currency: "Invalid",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "InternalServerError",
			owner:    account.Owner,
			currency: account.Currency,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(db.CreateAccountParams{
						Owner:    account.Owner,
						Balance:  0,
						Currency: account.Currency,
					})).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)

			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts"
			arg := createAccountRequest{
				Owner:    tc.owner,
				Currency: tc.currency,
			}
			data, err := json.Marshal(arg)
			require.NoError(t, err)

			request, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestDeleteAccountAPI(t *testing.T) {
	account := getRandomAccount()

	testCases := []struct {
		name          string
		id            int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			id:   account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				var body map[string]string
				err := json.Unmarshal(recorder.Body.Bytes(), &body)
				require.NoError(t, err)
				require.Equal(t, "account deleted", body["message"])
			},
		},
		{
			name: "NotFound",
			id:   account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(pgx.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "BadRequestFailValidation",
			id:   0,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			id:   account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					DeleteAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		ct := testCases[i]
		t.Run(ct.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			ct.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", ct.id)
			req, err := http.NewRequest("DELETE", url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			ct.checkResponse(t, recorder)
		})
	}
}

func TestListAccountsAPI(t *testing.T) {
	accounts := []db.Account{getRandomAccount(), getRandomAccount(), getRandomAccount()}

	testCases := []struct {
		name          string
		pageID        any
		pageSize      any
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:     "OK",
			pageID:   1,
			pageSize: 5, // valid per min=5,max=10
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{Limit: 5, Offset: 0})).
					Times(1).
					Return(accounts, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchAccounts(t, recorder.Body, accounts)
			},
		},
		{
			name:     "BadRequestMissingPageID",
			pageID:   "",
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "BadRequestMissingPageSize",
			pageID:   1,
			pageSize: "",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "BadRequestPageSizeTooSmall",
			pageID:   1,
			pageSize: 4,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "BadRequestPageSizeTooLarge",
			pageID:   1,
			pageSize: 11,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "BadRequestPageIDZero",
			pageID:   0,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:     "InternalServerError",
			pageID:   1,
			pageSize: 5,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					ListAccounts(gomock.Any(), gomock.Eq(db.ListAccountsParams{Limit: 5, Offset: 0})).
					Times(1).
					Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		ct := testCases[i]
		t.Run(ct.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			ct.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			// Build URL conditionally (omit params when we want missing)
			base := "/accounts"
			params := []string{}
			if v, ok := ct.pageID.(int); ok && v != 0 { // zero invalid triggers missing or bad depending on scenario
				params = append(params, fmt.Sprintf("page_id=%d", v))
			} else if v, ok := ct.pageID.(int32); ok && v != 0 {
				params = append(params, fmt.Sprintf("page_id=%d", v))
			} else if s, ok := ct.pageID.(string); ok && s != "" { // allow explicit string to simulate wrong type? here treat as provided
				params = append(params, fmt.Sprintf("page_id=%s", s))
			}
			if v, ok := ct.pageSize.(int); ok && v != 0 {
				params = append(params, fmt.Sprintf("page_size=%d", v))
			} else if v, ok := ct.pageSize.(int32); ok && v != 0 {
				params = append(params, fmt.Sprintf("page_size=%d", v))
			} else if s, ok := ct.pageSize.(string); ok && s != "" {
				params = append(params, fmt.Sprintf("page_size=%s", s))
			}
			url := base
			if len(params) > 0 {
				url = fmt.Sprintf("%s?%s", base, strings.Join(params, "&"))
			}

			req, err := http.NewRequest("GET", url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			ct.checkResponse(t, recorder)
		})
	}
}

func getRandomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requireBodyMatchAccount(t *testing.T, body *bytes.Buffer, account *db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, *account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, body *bytes.Buffer, accounts []db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	err = json.Unmarshal(data, &gotAccounts)
	require.NoError(t, err)
	require.Equal(t, accounts, gotAccounts)
}
