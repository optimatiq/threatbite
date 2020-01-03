package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigLists(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
		list    string
	}{
		{
			name:    "list default",
			wantErr: false,
		},
		{
			name:    "proxy list invalid one",
			wantErr: true,
			list:    "invalid_url",
		},
		{
			name:    "proxy list invalid one of",
			wantErr: true,
			list:    "https://some_url.com invalid_url",
		},
		{
			name:    "proxy list valid one",
			wantErr: false,
			list:    "https://some_url.com",
		},
		{
			name:    "proxy list valid more",
			wantErr: false,
			list:    "https://some_url.com https://next_url_.com",
		},
	}
	for _, env := range []string{"PROXY_LIST", "SPAM_LIST", "VPN_LIST", "DC_LIST", "EMAIL_DISPOSAL_LIST", "EMAIL_FREE_LIST"} {
		for _, tt := range tests {
			t.Run(tt.name+"_"+env, func(t *testing.T) {
				err := os.Setenv(env, tt.list)
				assert.NoError(t, err)

				_, err = NewConfig("")
				if (err != nil) != tt.wantErr {
					t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			})
		}
		os.Unsetenv(env)
	}
}
