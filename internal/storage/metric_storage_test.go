package storage

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestMapStorage_SaveCounter(t *testing.T) {
	type fields struct {
		gauges   map[string]float64
		counters map[string]int64
	}
	type args struct {
		name string
		val  int64
	}
	type want struct {
		expectedGaugeLen   int
		expectedCounterLen int
		expectedVal        int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			"check add new counter",
			fields{
				make(map[string]float64),
				make(map[string]int64),
			},
			args{
				"example",
				int64(1),
			},
			want{
				0,
				1,
				int64(1),
			},
		},
		{
			"check add existing counter",
			fields{
				make(map[string]float64),
				map[string]int64{
					"example": int64(1),
				},
			},
			args{
				"example",
				int64(1),
			},
			want{
				0,
				1,
				int64(2),
			},
		},
		{
			"check add new counter with existing",
			fields{
				make(map[string]float64),
				map[string]int64{
					"example1": int64(1),
				},
			},
			args{
				"example2",
				int64(1),
			},
			want{
				0,
				2,
				int64(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srm := &MapStorage{
				gauges:   tt.fields.gauges,
				counters: tt.fields.counters,
			}
			srm.SaveCounter(tt.args.name, tt.args.val)
			assert.Equal(t, tt.want.expectedGaugeLen, len(srm.gauges))
			assert.Equal(t, tt.want.expectedCounterLen, len(srm.counters))
		})
	}
}

func TestMapStorage_SaveGauge(t *testing.T) {
	type fields struct {
		gauges   map[string]float64
		counters map[string]int64
	}
	type args struct {
		name string
		val  float64
	}
	type want struct {
		expectedGaugeLen   int
		expectedCounterLen int
		expectedVal        float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			"check add new gauge",
			fields{
				make(map[string]float64),
				make(map[string]int64),
			},
			args{
				"example",
				float64(1),
			},
			want{
				1,
				0,
				float64(1),
			},
		},
		{
			"check add new gauge with existing",
			fields{
				map[string]float64{
					"example1": float64(1),
				},
				make(map[string]int64),
			},
			args{
				"example2",
				float64(1),
			},
			want{
				2,
				0,
				float64(1),
			},
		},
		{
			"check add existing gauge",
			fields{
				map[string]float64{
					"example": float64(1),
				},
				make(map[string]int64),
			},
			args{
				"example",
				float64(1),
			},
			want{
				1,
				0,
				float64(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srm := &MapStorage{
				gauges:   tt.fields.gauges,
				counters: tt.fields.counters,
			}
			srm.SaveGauge(tt.args.name, tt.args.val)
			assert.Equal(t, tt.want.expectedGaugeLen, len(srm.gauges))
			assert.Equal(t, tt.want.expectedCounterLen, len(srm.counters))
			assert.Equal(t, tt.want.expectedVal, srm.gauges[tt.args.name])
		})
	}
}

func Test_summWithCheck(t *testing.T) {
	tests := []struct {
		var1          int64
		var2          int64
		expectedError bool
	}{
		{
			1, 2, false,
		},
		{
			0, 0, false,
		},
		{
			math.MaxInt64, 0, false,
		},
		{
			math.MaxInt64, 1, true,
		},
		{
			math.MaxInt64 - 2, 2, false,
		},
		{
			math.MaxInt64, math.MaxInt64, true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("var1=%d,var2=%d", tt.var1, tt.var2), func(t *testing.T) {
			_, e := sumWithCheck(tt.var1, tt.var2)
			assert.Equal(t, tt.expectedError, e != nil)
		})
	}
}
