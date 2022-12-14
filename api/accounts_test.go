package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/orlandorode97/simple-bank/generated/sql/simplebanksql"
	"github.com/orlandorode97/simple-bank/store/mockdb"
)

type accountResponse struct {
	Accounts []simplebanksql.Account `json:"accounts"`
}

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		desc                string
		request             createAccountRequest
		createAccountTimes  int
		createAccountParams simplebanksql.CreateAccountParams
		createAccountErr    error

		wantAccount  simplebanksql.Account
		wantHTTPCode int
	}{
		{
			desc: "success - account was created",
			request: createAccountRequest{
				Owner:      "Joe Sample",
				CurrencyID: 1,
			},
			createAccountTimes: 1,
			createAccountParams: simplebanksql.CreateAccountParams{
				Owner:      "Joe Sample",
				Balance:    0,
				CurrencyID: 1,
			},

			wantAccount: simplebanksql.Account{
				ID:      1,
				Owner:   "Joe Sample",
				Balance: 0,
			},
			wantHTTPCode: http.StatusCreated,
		},
		{
			desc: "failure - missing owner",
			request: createAccountRequest{
				CurrencyID: 1,
			},

			wantHTTPCode: http.StatusBadRequest,
		},
		{
			desc: "failure - missing currency id",
			request: createAccountRequest{
				Owner: "Chuck Loeb",
			},

			wantHTTPCode: http.StatusBadRequest,
		},
		{
			desc: "failure - unable to create account",
			request: createAccountRequest{
				Owner:      "Chuck Loeb",
				CurrencyID: 1,
			},
			createAccountTimes: 1,
			createAccountParams: simplebanksql.CreateAccountParams{
				Owner:      "Chuck Loeb",
				Balance:    0,
				CurrencyID: 1,
			},
			createAccountErr: sql.ErrConnDone,

			wantHTTPCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/api/v1/accounts")
			store.EXPECT().
				CreateAccount(gomock.Any(), gomock.Eq(tc.createAccountParams)).
				Times(tc.createAccountTimes).
				Return(tc.wantAccount, tc.createAccountErr)

			payload, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatal(err)
			}

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
			if err != nil {
				t.Fatal(err)
			}

			server.handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.wantHTTPCode {
				t.Fatalf("got %v but expected %v", recorder.Code, tc.wantHTTPCode)
			}
		})
	}
}

func TestGetAccount(t *testing.T) {
	tests := []struct {
		desc            string
		request         getAccountRequest
		getAccountTimes int
		getAccountErr   error

		wantAccount  simplebanksql.Account
		wantHTTPCode int
	}{
		{
			desc: "success - account is returned",
			request: getAccountRequest{
				ID: 1,
			},
			getAccountTimes: 1,

			wantAccount: simplebanksql.Account{
				ID:      1,
				Owner:   "Joe Sample",
				Balance: 0,
			},
			wantHTTPCode: http.StatusOK,
		},
		{
			desc: "failure - account id is less than default (1)",
			request: getAccountRequest{
				ID: -1,
			},
			getAccountTimes: 0,

			wantHTTPCode: http.StatusBadRequest,
		},
		{
			desc: "failure - account does not exist",
			request: getAccountRequest{
				ID: 1,
			},
			getAccountTimes: 1,
			getAccountErr:   sql.ErrNoRows,

			wantHTTPCode: http.StatusNotFound,
		},
		{
			desc: "failure - database is not avaialable",
			request: getAccountRequest{
				ID: 1,
			},
			getAccountTimes: 1,
			getAccountErr:   sql.ErrConnDone,

			wantHTTPCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/api/v1/accounts/%v", tc.request.ID)

			store.EXPECT().
				GetAccount(gomock.Any(), gomock.Eq(tc.request.ID)).
				Times(tc.getAccountTimes).
				Return(tc.wantAccount, tc.getAccountErr)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			server.handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.wantHTTPCode {
				t.Fatalf("got %v but expected %v", recorder.Code, tc.wantHTTPCode)
			}
			var account simplebanksql.Account
			body, err := io.ReadAll(recorder.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(body, &account); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(account, tc.wantAccount) {
				t.Fatalf("got %v but expected %v", account, tc.wantAccount)
			}
		})
	}
}

func TestListAccounts(t *testing.T) {
	tests := []struct {
		desc               string
		request            listAccountsRequest
		listAccountTimes   int
		listAccountParams  simplebanksql.ListAccountsParams
		listAccountErr     error
		listAccountsReturn []simplebanksql.Account

		wantAccounts accountResponse
		wantHTTPCode int
	}{
		{
			desc: "success - list of accounts is returned",
			request: listAccountsRequest{
				PageID:   1,
				PageSize: 5,
			},
			listAccountTimes: 1,
			listAccountParams: simplebanksql.ListAccountsParams{
				Limit:  5,
				Offset: 0,
			},
			listAccountsReturn: []simplebanksql.Account{
				{
					ID:      1,
					Owner:   "Bob James",
					Balance: 1,
				},
			},

			wantAccounts: accountResponse{
				Accounts: []simplebanksql.Account{
					{
						ID:      1,
						Owner:   "Bob James",
						Balance: 1,
					},
				},
			},
			wantHTTPCode: http.StatusOK,
		},
		{
			desc: "success - list of accounts is empty",
			request: listAccountsRequest{
				PageID:   1,
				PageSize: 5,
			},
			listAccountTimes: 1,
			listAccountParams: simplebanksql.ListAccountsParams{
				Limit:  5,
				Offset: 0,
			},
			listAccountsReturn: []simplebanksql.Account{},

			wantAccounts: accountResponse{
				Accounts: []simplebanksql.Account{},
			},
			wantHTTPCode: http.StatusOK,
		},
		{
			desc: "failure - page id is less than default (1)",
			request: listAccountsRequest{
				PageID:   0,
				PageSize: 5,
			},

			wantHTTPCode: http.StatusBadRequest,
		},
		{
			desc: "failure - page size is less than default (5)",
			request: listAccountsRequest{
				PageID:   1,
				PageSize: 4,
			},

			wantHTTPCode: http.StatusBadRequest,
		},
		{
			desc: "failure - unable to connect to database",
			request: listAccountsRequest{
				PageID:   1,
				PageSize: 5,
			},
			listAccountTimes: 1,
			listAccountParams: simplebanksql.ListAccountsParams{
				Limit:  5,
				Offset: 0,
			},
			listAccountsReturn: []simplebanksql.Account{},
			listAccountErr:     sql.ErrConnDone,

			wantHTTPCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			server := NewServer(store)
			recorder := httptest.NewRecorder()
			url := fmt.Sprintf("/api/v1/accounts?page_id=%v&page_size=%v", tc.request.PageID, tc.request.PageSize)

			store.EXPECT().
				ListAccounts(gomock.Any(), tc.listAccountParams).
				Times(tc.listAccountTimes).
				Return(tc.listAccountsReturn, tc.listAccountErr)

			request, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Fatal(err)
			}

			server.handler.ServeHTTP(recorder, request)
			if recorder.Code != tc.wantHTTPCode {
				t.Fatalf("got %v but expected %v", recorder.Code, tc.wantHTTPCode)
			}

			var accounts accountResponse
			body, err := io.ReadAll(recorder.Body)
			if err != nil {
				t.Fatal(err)
			}

			if err = json.Unmarshal(body, &accounts); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(accounts, tc.wantAccounts) {
				t.Fatalf("got %v but expected %v", accounts, tc.wantAccounts)
			}
		})
	}
}
