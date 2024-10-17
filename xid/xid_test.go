package xid

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ceeideu/sdk/properties"
)

func Test_val(t *testing.T) {
	t.Parallel()
	type args struct {
		version byte
		t       TypeOf
		fn      func([]byte) error
	}
	tests := []struct {
		name    string
		args    args
		want    Value
		wantErr bool
	}{
		{
			name: "xid",
			args: args{
				version: 0xA,
				t:       TypeOf(0xB),
				fn: func(b []byte) error {
					for i := range b {
						b[i] = 0xf
					}

					return nil
				},
			},
			want: Value{0xA /* version */, 0xB /* unknown type */, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf, 0xf},
		},
		{
			name: "err",
			args: args{
				version: 0xA,
				t:       TypeOf(0xB),
				fn: func(b []byte) error {
					return fmt.Errorf("%s", "err")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := val(test.args.version, test.args.t, test.args.fn)
			if (err != nil) != test.wantErr {
				t.Errorf("val() error = %v, wantErr %v", err, test.wantErr)

				return
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("val() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestTypeOf_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		tr   TypeOf
		want string
	}{
		{
			name: "unknown",
			tr:   TypeOfUnknown,
			want: "unknown",
		},
		{
			name: "email",
			tr:   TypeOfEmail,
			want: "email",
		},
		{
			name: "foo",
			tr:   TypeOf(0xF),
			want: "unknown",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.tr.String(); got != test.want {
				t.Errorf("TypeOf.String() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestValue_Type(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		x    Value
		want byte
	}{
		{
			name: "typ",
			x:    Value{0xF, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB, 0xB},
			want: 0xB,
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.x.Type(); got != test.want {
				t.Errorf("Value.Type() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRefreshReq_WithProperties(t *testing.T) {
	t.Parallel()
	type fields struct {
		XID        string
		Properties map[string]string
	}
	type args struct {
		_properties properties.Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   RefreshReq
	}{
		{
			name: "1",
			args: args{_properties: properties.WithConsent("foo")},
			want: RefreshReq{Properties: map[string]string{"consent": "foo"}},
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			r := RefreshReq{
				XID:        test.fields.XID,
				Properties: test.fields.Properties,
			}
			if got := r.WithProperties(test.args._properties); !reflect.DeepEqual(got, test.want) {
				t.Errorf("RefreshReq.WithProperties() = %v, want %v", got, test.want)
			}
		})
	}
}
