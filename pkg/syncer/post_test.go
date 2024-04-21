package syncer

import "testing"

func Test_getJson(t *testing.T) {
	type args struct {
		u string
		v interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				u: "https://konachan.com/tag.json?limit=10&order=count",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if str, err := getJson(tt.args.u); (err != nil) != tt.wantErr {
				t.Errorf("getJson() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("%v", str)
			}
		})
	}
}
