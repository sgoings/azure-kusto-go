package azkustodata

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/tj/assert"
)

func TestGetConnectionStringBuilder(t *testing.T) {

	tests := []struct {
		name             string
		connectionString string
		want             ConnectionStringBuilder
		wantErr          string
	}{
		{
			name:             "test_conn_string_validURL",
			connectionString: "https://endpoint",
			want: ConnectionStringBuilder{
				DataSource: "https://endpoint",
			},
		},
		{
			name:             "test_conn_string_emptyconnstr",
			connectionString: "",
			wantErr:          "error : Connection string cannot be empty",
		},
		{
			name:             "test_conn_string_fullstring",
			connectionString: "https://help.kusto.windows.net/Samples;aad user id=1234;password=****;application key=1234;application client id=1234;application key=0987;application certificate=avsefsfbsrgbrb; authority id=123456;application token=token;user token=usertoken;;interactivelogin=false; domainhint=www.google.com",
			want: ConnectionStringBuilder{
				DataSource:                 "https://help.kusto.windows.net/Samples",
				AadUserID:                  "1234",
				Password:                   "****",
				UserToken:                  "usertoken",
				ApplicationClientId:        "1234",
				ApplicationKey:             "0987",
				AuthorityId:                "123456",
				ApplicationCertificatePath: "avsefsfbsrgbrb",
				SendCertificateChain:       false,
				ApplicationToken:           "token",
				AzCli:                      false,
				MsiAuthentication:          false,
				ManagedServiceIdentity:     "",
				InteractiveLogin:           false,
				RedirectURL:                "www.google.com",
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if isEmpty(test.wantErr) {
				actual := NewConnectionStringBuilder(test.connectionString)
				actual.ApplicationForTracing = ""
				actual.UserForTracing = ""
				assert.EqualValues(t, test.want, *actual)
			} else {
				assert.Panics(t, func() { NewConnectionStringBuilder(test.connectionString) }, test.wantErr)
			}
		})

	}
}

func TestWithAadUserPassAuth(t *testing.T) {
	want := ConnectionStringBuilder{
		DataSource:  "endpoint",
		AadUserID:   "userid",
		Password:    "password",
		AuthorityId: "authorityID",
	}

	actual := NewConnectionStringBuilder("endpoint").WithAadUserPassAuth("userid", "password", "authorityID")
	actual.ApplicationForTracing = ""
	actual.UserForTracing = ""

	assert.EqualValues(t, want, *actual)
}

func TestWithAadUserPassAuthErr(t *testing.T) {
	defer func() {
		if res := recover(); res == nil {
			t.Errorf("Should have panic")
		} else if res != "Error: Password cannot be null" {
			t.Errorf("Wrong panic message: %s", res)
		}
	}()
	NewConnectionStringBuilder("endpoint").WithAadUserPassAuth("userid", "", "authorityID")

}

func TestWitAadUserToken(t *testing.T) {
	want := ConnectionStringBuilder{
		DataSource: "endpoint",
		UserToken:  "token",
	}

	actual := NewConnectionStringBuilder("endpoint").WitAadUserToken("token")
	actual.ApplicationForTracing = ""
	actual.UserForTracing = ""
	assert.EqualValues(t, want, *actual)
}

func TestWithWorkloadIdentity(t *testing.T) {
	want := ConnectionStringBuilder{
		DataSource:              "endpoint",
		ApplicationClientId:     "clientID",
		AuthorityId:             "authorityID",
		FederationTokenFilePath: "tokenfilepath",
		WorkloadAuthentication:  true,
	}

	actual := NewConnectionStringBuilder("endpoint").WithKubernetesWorkloadIdentity("clientID", "tokenfilepath", "authorityID")

	assert.EqualValues(t, want, *actual)
}

func TestWitAadUserTokenErr(t *testing.T) {
	defer func() {
		if res := recover(); res == nil {
			t.Errorf("Should have panic")
		} else if res != "Error: UserToken cannot be null" {
			t.Errorf("Wrong panic message: %s", res)
		}
	}()
	NewConnectionStringBuilder("endpoint").WitAadUserToken("")

}

func TestGetTokenProviderHappy(t *testing.T) {
	tests := []struct {
		name    string
		kcsb    ConnectionStringBuilder
		payload string
	}{
		{
			name: "test_tokenprovider_usernamepasswordauth",
			kcsb: ConnectionStringBuilder{
				DataSource:          "https://endpoint/test_tokenprovider_usernamepasswordauth",
				AuthorityId:         "tenantID",
				ApplicationClientId: "clientID",
				AadUserID:           "ussername",
				Password:            "userpass",
			},
		}, {
			name: "test_tokenprovider_intLogin",
			kcsb: ConnectionStringBuilder{
				DataSource:          "https://endpoint/test_tokenprovider_intLogin",
				InteractiveLogin:    true,
				AuthorityId:         "tenantID",
				ApplicationClientId: "clientID",
			},
		},
		{
			name: "test_tokenprovider_clientsec",
			kcsb: ConnectionStringBuilder{
				DataSource:          "https://endpoint/test_tokenprovider_clientsec",
				InteractiveLogin:    true,
				AuthorityId:         "tenantID",
				ApplicationClientId: "clientID",
				ApplicationKey:      "somekey",
			},
		}, {
			name: "test_tokenprovider_managedsi",
			kcsb: ConnectionStringBuilder{
				DataSource:             "https://endpoint/test_tokenprovider_managedsi",
				ManagedServiceIdentity: "managedid",
				MsiAuthentication:      true,
				ClientOptions:          &azcore.ClientOptions{},
			},
		}, {
			name: "test_tokenprovider_managedidauth2",
			kcsb: ConnectionStringBuilder{
				DataSource:        "https://endpoint/test_tokenprovider_managedidauth2",
				MsiAuthentication: true,
			},
		}, {
			name: "test_tokenprovider_workloadidentity",
			kcsb: ConnectionStringBuilder{
				DataSource:              "https://endpoint/test_tokenprovider_workloadidentity",
				ApplicationClientId:     "clientID",
				AuthorityId:             "tenantID",
				FederationTokenFilePath: "tokenfilepath",
				WorkloadAuthentication:  true,
			},
		}, {
			name: "test_tokenprovider_usertoken",
			kcsb: ConnectionStringBuilder{
				DataSource: "https://endpoint/test_tokenprovider_usertoken",
				UserToken:  "token",
			},
		}, {
			name: "test_tokenprovider_apptoken",
			kcsb: ConnectionStringBuilder{
				DataSource: "https://endpoint/test_tokenprovider_apptoken",
				UserToken:  "token",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			kscb := test.kcsb
			got, err := kscb.newTokenProvider()
			assert.Nil(t, err)
			assert.NotNil(t, got)
		})
	}

}
