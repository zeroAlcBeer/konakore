package client

import (
	"fmt"
	"testing"
)

var reqclient = New()

func TestReqClient_Download(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				url: "https://konachan.com/image/02d8a92202a4f723fa5f381451100db3/1.png",
			},
		},
		{
			name: "test",
			args: args{
				url: "https://konachan.com/image/02d8a92202a4f723fa5f381451100db3/1.jpg",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test, err := reqclient.CheckDownloadUrl(tt.args.url)
			if (err != nil) != tt.wantErr {

				t.Errorf("Download() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(test)
		})
	}
}
