package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
	mockdb "github.com/roman-adamchik/simplebank/db/mock"
	db "github.com/roman-adamchik/simplebank/db/sqlc"
	"github.com/roman-adamchik/simplebank/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTransferAPI(t *testing.T) {
	fromAccount := getRandomAccount()
	toAccount := getRandomAccount()
	currency := fromAccount.Currency
	toAccount.Currency = currency

	amount := util.RandomMoney()
	if amount <= 0 {
		amount = 1
	}

	transferResult := db.TransferTxResult{
		Transfer: db.Transfer{
			ID:            util.RandomInt(1, 1000),
			FromAccountID: fromAccount.ID,
			ToAccountID:   toAccount.ID,
			Amount:        amount,
		},
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		FromEntry: db.Entry{
			ID:        util.RandomInt(1, 1000),
			AccountID: fromAccount.ID,
			Amount:    -amount,
		},
		ToEntry: db.Entry{
			ID:        util.RandomInt(1, 1000),
			AccountID: toAccount.ID,
			Amount:    amount,
		},
	}

	testCases := []struct {
		name          string
		body          transferRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: fromAccount.ID,
						ToAccountID:   toAccount.ID,
						Amount:        amount,
					})).
					Times(1).
					Return(transferResult, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTransferResult(t, recorder.Body, &transferResult)
			},
		},
		{
			name: "BadRequestFromAccountIDMissing",
			body: transferRequest{
				FromAccountID: 0,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestToAccountIDMissing",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   0,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestAmountZero",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        0,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestAmountNegative",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        -1,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestCurrencyInvalid",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      "INVALID",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestCurrencyMissing",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      "",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "NotFoundFromAccount",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "NotFoundToAccount",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrNoRows)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "BadRequestCurrencyMismatchFromAccount",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      "USD",
			},
			buildStubs: func(store *mockdb.MockStore) {
				// Set fromAccount to have a different currency
				fromAccountWithDifferentCurrency := fromAccount
				fromAccountWithDifferentCurrency.Currency = "EUR"
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccountWithDifferentCurrency, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "BadRequestCurrencyMismatchToAccount",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				// Set toAccount to have a different currency
				toAccountWithDifferentCurrency := toAccount
				toAccountWithDifferentCurrency.Currency = "ILS"
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccountWithDifferentCurrency, nil)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InternalServerErrorFromGetAccount",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(db.Account{}, pgx.ErrTxClosed)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(0)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InternalServerErrorFromTransferTx",
			body: transferRequest{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
				Currency:      currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(fromAccount.ID)).
					Times(1).
					Return(fromAccount, nil)
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(toAccount.ID)).
					Times(1).
					Return(toAccount, nil)
				store.EXPECT().
					TransferTx(gomock.Any(), gomock.Eq(db.TransferTxParams{
						FromAccountID: fromAccount.ID,
						ToAccountID:   toAccount.ID,
						Amount:        amount,
					})).
					Times(1).
					Return(db.TransferTxResult{}, pgx.ErrTxClosed)
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

			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest("POST", "/transfers", bytes.NewBuffer(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchTransferResult(t *testing.T, body *bytes.Buffer, result *db.TransferTxResult) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotResult db.TransferTxResult
	err = json.Unmarshal(data, &gotResult)
	require.NoError(t, err)
	require.Equal(t, result.Transfer, gotResult.Transfer)
	require.Equal(t, result.FromAccount, gotResult.FromAccount)
	require.Equal(t, result.ToAccount, gotResult.ToAccount)
	require.Equal(t, result.FromEntry, gotResult.FromEntry)
	require.Equal(t, result.ToEntry, gotResult.ToEntry)
}
