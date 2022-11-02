package handler

import "testing"

func Test_createShortenURL(t *testing.T) {
	type args struct {
		id      string
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "positive1",
			args: args{
				"1",
				"http://localhost:8080",
			},
			want: "http://localhost:8080/1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createShortenURL(tt.args.id, tt.args.baseURL); got != tt.want {
				t.Errorf("createShortenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
