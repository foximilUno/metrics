package handlers

import (
	"github.com/foximilUno/metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSaveMetrics(t *testing.T) {
	type args struct {
		method      string
		contentType string
		url         string
	}
	type want struct {
		expectedStatucCode int
		expectedBody       string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"test service check method",
			args{
				http.MethodGet,
				"",
				"",
			},
			want{
				405,
				"Only POST allowed\n",
			},
		},
		//TODO if check content type - fails yandex-practicum tests
		//{
		//	"test content type check",
		//	args{
		//		http.MethodPost,
		//		"application/json",
		//		"/update/gauge/metric/1",
		//	},
		//	want{
		//		415,
		//		"Allowed text/plain only\n",
		//	},
		//},
		{
			"test path check without update",
			args{
				http.MethodPost,
				"text/plain",
				"",
			},
			want{
				400,
				"path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>\n",
			},
		},
		{
			"test path check with only update",
			args{
				http.MethodPost,
				"text/plain",
				"/update",
			},
			want{
				400,
				"path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>\n",
			},
		},
		{
			"test path with wrong number of params",
			args{
				http.MethodPost,
				"text/plain",
				"/update/gauge/1/2/3",
			},
			want{
				400,
				"path must be pattern like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>\n",
			},
		},
		{
			"test success gauge",
			args{
				http.MethodPost,
				"text/plain",
				"/update/gauge/1/2",
			},
			want{
				200,
				"",
			},
		},
		{
			"test success conter",
			args{
				http.MethodPost,
				"text/plain",
				"/update/counter/1/2",
			},
			want{
				200,
				"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.args.method, tt.args.url, nil)
			assert.NoError(t, err)
			req.Header.Set("Content-type", tt.args.contentType)
			w := httptest.NewRecorder()
			h := SaveMetricsViaTextPlain(storage.NewMapStorage())
			h.ServeHTTP(w, req)
			r := w.Result()
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			defer r.Body.Close()
			assert.Equal(t, tt.want.expectedStatucCode, r.StatusCode)
			assert.Equal(t, tt.want.expectedBody, string(body))
		})
	}
}
