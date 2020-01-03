package email

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/optimatiq/threatbite/email/datasource"
)

func Test_free_isFree(t *testing.T) {
	type fields struct {
		domain *datasource.Domain
	}
	type args struct {
		email string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "on list",
			fields: fields{
				domain: datasource.NewDomain(datasource.NewListDataSource([]string{"gmail.com"}), "free"),
			},
			args: args{
				email: "xxx@gmail.com",
			},
			want: true,
		},
		{
			name: "on list 2",
			fields: fields{
				domain: datasource.NewDomain(datasource.NewListDataSource([]string{"xxx.com", "gmail.com"}), "free"),
			},
			args: args{
				email: "xxx@gmail.com",
			},
			want: true,
		},
		{
			name: "on list case sensitive",
			fields: fields{
				domain: datasource.NewDomain(datasource.NewListDataSource([]string{"GMAIL.com"}), "free"),
			},
			args: args{
				email: "xxx@gmail.COM",
			},
			want: true,
		},
		{
			name: "not on list",
			fields: fields{
				domain: datasource.NewDomain(datasource.NewListDataSource([]string{"zzz.com"}), "free"),
			},
			args: args{
				email: "xxx@xxx.com",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &free{
				domain: tt.fields.domain,
			}

			err := d.domain.Load()
			assert.NoError(t, err)

			if got := d.isFree(tt.args.email); got != tt.want {
				t.Errorf("isFree() = %v, want %v", got, tt.want)
			}
		})
	}
}
