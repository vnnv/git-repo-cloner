package main

import "testing"

const sshUser = "git"

func Test_generateSshRemoteUrl(t *testing.T) {
	type args struct {
		httpUrl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{name: "happyPath_https", args: args{"https://github.com/vnnv/go-epay.git"}, want: "git@github.com:vnnv/go-epay.git"},
		{name: "happyPath_http", args: args{"http://github.com/vnnv/go-epay.git"}, want: "git@github.com:vnnv/go-epay.git"},
		{name: "happyPath_ftp", args: args{"ftp://github.com/vnnv/go-epay.git"}, want: "git@github.com:vnnv/go-epay.git"},
		{name: "happyPath_no_proto", args: args{"github.com/vnnv/go-epay.git"}, want: "git@github.com:vnnv/go-epay.git"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateSshRemoteUrl(tt.args.httpUrl, sshUser); got != tt.want {
				t.Errorf("generateSshRemoteUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}