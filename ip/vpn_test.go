package ip

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimatiq/threatbite/ip/datasource"
)

func Test_vpn_isVpn(t *testing.T) {
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name    string
		list    []string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "on list",
			list: []string{"192.168.0.1"},
			args: args{
				ip: net.ParseIP("192.168.0.1"),
			},
			want: true,
		},
		{
			name: "on list cidr",
			list: []string{"192.168.0.0/24"},
			args: args{
				ip: net.ParseIP("192.168.0.1"),
			},
			want: true,
		},
		{
			name: "not on list",
			list: []string{"192.168.0.0/24"},
			args: args{
				ip: net.ParseIP("191.168.0.1"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := datasource.NewListDataSource(tt.list)
			assert.NoError(t, err)
			v := &vpn{
				ipnet: datasource.NewIPNet(ds, "vpn"),
			}
			err = v.ipnet.Load()
			assert.NoError(t, err)

			got, err := v.isVpn(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("isVpn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isVpn() got = %v, want %v", got, tt.want)
			}
		})
	}
}
