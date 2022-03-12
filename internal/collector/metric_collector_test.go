package collector

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func Test_collector_addGauge(t *testing.T) {
	type fields struct {
		baseURL string
		data    map[string]*MetricEntity
	}
	type args struct {
		name  string
		value uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"check add new gauge",
			fields{
				"",
				make(map[string]*MetricEntity),
			},
			args{
				"val",
				uint64(1),
			},
		},
		{
			"check add existing gauge",
			fields{
				"",
				map[string]*MetricEntity{
					"val": {
						"gauge",
						"val",
						uint64(1),
					}},
			},
			args{
				"val",
				uint64(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &collector{
				baseURL: tt.fields.baseURL,
				data:    tt.fields.data,
			}
			mc.addGauge(tt.args.name, tt.args.value)
			assert.Equal(t, tt.args.value, mc.data[tt.args.name].entityValue)
			assert.Equal(t, "gauge", mc.data[tt.args.name].entityType)
		})

	}
}

func Test_collector_increaseCounter(t *testing.T) {
	type fields struct {
		baseURL string
		data    map[string]*MetricEntity
	}
	type args struct {
		name        string
		expectedVal uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			"check add new counter",
			fields{
				"",
				make(map[string]*MetricEntity),
			},
			args{
				"val",
				uint64(1),
			},
		},
		{
			"check add existing counter",
			fields{
				"",
				map[string]*MetricEntity{
					"val": {
						"counter",
						"val",
						uint64(1),
					},
				},
			},
			args{
				"val",
				uint64(2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &collector{
				baseURL: tt.fields.baseURL,
				data:    tt.fields.data,
			}
			mc.increaseCounter(tt.args.name)
			assert.Equal(t, tt.args.expectedVal, mc.data[tt.args.name].entityValue)
			assert.Equal(t, "counter", mc.data[tt.args.name].entityType)
		})
	}
}

func Test_collector_getEntity(t *testing.T) {
	type fields struct {
		baseURL string
		data    map[string]*MetricEntity
	}
	type args struct {
		name        string
		typeentity  string
		expectedLen int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *MetricEntity
	}{
		{
			"check get not existing gauge entity",
			fields{
				"",
				make(map[string]*MetricEntity),
			},
			args{
				"val",
				"gauge",
				1,
			},
			&MetricEntity{
				"gauge",
				"val",
				uint64(0),
			},
		},
		{
			"check get existing gauge entity",
			fields{
				"",
				map[string]*MetricEntity{
					"val": {
						"gauge",
						"val",
						uint64(0),
					},
				},
			},
			args{
				"val",
				"gauge",
				1,
			},
			&MetricEntity{
				"gauge",
				"val",
				uint64(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &collector{
				baseURL: tt.fields.baseURL,
				data:    tt.fields.data,
			}
			if got := mc.getEntity(tt.args.name, tt.args.typeentity); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEntity() = %v, want %v", got, tt.want)
			}
			assert.Equal(t, tt.args.expectedLen, len(mc.data))
		})
	}
}

func Test_collector_Collect(t *testing.T) {
	type fields struct {
		baseURL string
		data    map[string]*MetricEntity
	}
	tests := []struct {
		name              string
		fields            fields
		invokeRetryNumber int
		expectedLen       int
	}{
		{
			"check first run fulled ",
			fields{
				"",
				make(map[string]*MetricEntity),
			},
			1,
			29,
		},
		{
			"check two runs filled map",
			fields{
				"",
				make(map[string]*MetricEntity),
			},
			2,
			29,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &collector{
				baseURL: tt.fields.baseURL,
				data:    tt.fields.data,
			}
			for i := 0; i < tt.invokeRetryNumber; i++ {
				mc.Collect()
			}
			assert.Equal(t, tt.expectedLen, len(mc.data))
		})
	}
}
