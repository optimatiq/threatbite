package ip

import (
	"net"
	"testing"

	"github.com/optimatiq/threatbite/ip/datasource"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/mock"
)

type mockedGeoip struct {
	mock.Mock
}

func (m *mockedGeoip) getCountry(ip net.IP) (string, error) {
	args := m.Called(ip)
	return args.String(0), args.Error(1)
}

func (m *mockedGeoip) getCompany(ip net.IP) (string, error) {
	args := m.Called(ip)
	return args.String(0), args.Error(1)
}

func (m *mockedGeoip) update() error {
	return nil
}

func Test_datacenter_isDC(t *testing.T) {
	geo := new(mockedGeoip)
	geo.On("getCountry", net.ParseIP("1.1.1.1")).Return("PL", nil)
	geo.On("getCompany", net.ParseIP("1.1.1.1")).Return("Misc corp.", nil)

	geo.On("getCountry", net.ParseIP("1.1.1.2")).Return("PL", nil)
	geo.On("getCompany", net.ParseIP("1.1.1.2")).Return("Misc corp.", nil)

	geo.On("getCountry", net.ParseIP("1.1.1.3")).Return("PL", nil)
	geo.On("getCompany", net.ParseIP("1.1.1.3")).Return("OVH corporation", nil)

	type fields struct {
		list  []string
		geoip geoip
	}
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "on list",
			fields: fields{
				list:  []string{"1.1.1.1"},
				geoip: geo,
			},
			args: args{ip: net.ParseIP("1.1.1.1")},
			want: true,
		},

		{
			name: "not on list",
			fields: fields{
				list:  []string{"1.1.1.1"},
				geoip: geo,
			},
			args: args{ip: net.ParseIP("1.1.1.2")},
			want: false,
		},

		{
			name: "invalid IP",
			fields: fields{
				list:  []string{},
				geoip: geo,
			},
			args:    args{ip: net.ParseIP("invalid")},
			wantErr: true,
		},

		{
			name: "company on list",
			fields: fields{
				list:  []string{"1.1.1.1"},
				geoip: geo,
			},
			args: args{ip: net.ParseIP("1.1.1.3")},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds, err := datasource.NewListDataSource(tt.fields.list)
			assert.NoError(t, err)
			d := &datacenter{
				ipnet: datasource.NewIPNet(ds, "vpn"),
				geoip: tt.fields.geoip,
			}
			err = d.ipnet.Load()
			assert.NoError(t, err)

			got, err := d.isDC(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("isDC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isDC() got = %v, want %v", got, tt.want)
			}
		})
	}
}
