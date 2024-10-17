package hem

import (
	"reflect"
	"testing"

	"github.com/ceeideu/sdk/properties"
	"github.com/ceeideu/sdk/xid"
)

func TestFromEmail(t *testing.T) {
	t.Parallel()
	type args struct {
		email string
	}
	tests := []struct {
		name string
		args args
		want Request
	}{
		{
			name: "base",
			args: args{
				email: "alama@kota.pl",
			},
			want: Request{
				Type:  xid.Email,
				Value: "MJTGXA3+NSOZ9YMT0UOP8HhJfo76zzaKf52RiaKL/7c=",
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := FromEmail(test.args.email); !reflect.DeepEqual(got, test.want) {
				t.Errorf("[%s] FromEmail() = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		email   string
		want    string
		wantErr bool
	}{
		{
			email:   "",
			want:    "",
			wantErr: true,
		},
		{
			email:   "\tfoo@bar.baz\t",
			want:    "foo@bar.baz",
			wantErr: false,
		},
		{
			email:   "\tfoo+bar@baz.pl\t",
			want:    "foo@baz.pl",
			wantErr: false,
		},
		{
			email:   "\tfoo+bar++@baz.pl\t",
			want:    "foo@baz.pl",
			wantErr: false,
		},
		{
			email:   "randomstring",
			want:    "",
			wantErr: true,
		},
		{
			email:   "a.b.c.d@gmail.com",
			want:    "abcd@gmail.com",
			wantErr: false,
		},
		{
			email:   "a.b.c.d@test",
			want:    "a.b.c.d@test",
			wantErr: false,
		},
		{
			email:   "a.b.c.d+str@gmail.com",
			want:    "abcd@gmail.com",
			wantErr: false,
		},
		{
			email:   "AbCd@baz.pL",
			want:    "abcd@baz.pl",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.email, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeEmail(test.email)
			if (err != nil) != test.wantErr {
				t.Errorf("NormalizeEmail() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if got != test.want {
				t.Errorf("NormalizeEmail() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestFromHex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		hem     string
		wantErr bool
		want    Request
	}{
		{
			hem:     "foo",
			wantErr: true,
		},
		{
			hem:     "fffffffffffffffffffffffffffffffffffffffffffffffffffzzzzzzzzzzzzz",
			wantErr: true,
		},
		{
			hem:     "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			want:    Request{Type: xid.Hex, Value: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := FromHex(test.hem)

			if (got.Err != nil) != test.wantErr {
				t.Errorf("FromHex() error = %v, wantErr %v", got.Err, test.wantErr)

				return
			}

			if got.Value != test.want.Value || got.Type != test.want.Type {
				t.Errorf("FromHex() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRequest_WithProperties(t *testing.T) {
	t.Parallel()
	type fields struct {
		Type       string
		Value      string
		Err        error
		Properties map[string]string
	}
	type args struct {
		_properties properties.Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Request
	}{
		{
			name: "1",
			args: args{
				_properties: properties.WithConsent("foo"),
			},
			want: Request{
				Properties: map[string]string{"consent": "foo"},
			},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			request := Request{
				Type:       test.fields.Type,
				Value:      test.fields.Value,
				Err:        test.fields.Err,
				Properties: test.fields.Properties,
			}
			if got := request.WithProperties(test.args._properties); !reflect.DeepEqual(got, test.want) {
				t.Errorf("Request.WithProperties() = %v, want %v", got, test.want)
			}
		})
	}
}
