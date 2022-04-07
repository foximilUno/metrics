package secure

import (
	"github.com/foximilUno/metrics/internal/types"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestEncryptGaugeMetric(t *testing.T) {
	f := float64(32768)

	type args struct {
		m   *types.Metrics
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "check hash",
			args: args{
				m: &types.Metrics{
					ID:    "MSpanSys",
					MType: "gauge",
					Value: &f,
				},
				key: "/tmp/SuIBy",
			},
			want: "5af4af330889cf92ca53b6af9894918f974472e2199151aba2eba3d022c75fef",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptMetric(tt.args.m, tt.args.key)
			if err != nil {
				t.Errorf("EncryptMetric() : %e", err)
			}
			if got != tt.want {
				t.Errorf("EncryptMetric()  \n\t\tgot = %v, \n\t\twant %v", got, tt.want)
			}
		})
	}
}

func TestDecryptMetric(t *testing.T) {
	msg := strings.Repeat("bla", 10)
	key := "/tmp/SuIBy"
	h, err := encryptString(msg, key)
	assert.NoError(t, err)

	isEqual, err := IsValidHash(msg, h, key)
	assert.NoError(t, err)
	assert.Truef(t, isEqual, "isEqual не равен true")
}
