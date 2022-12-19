package gisha

import (
	"testing"
)

func Test_getRandomAvailablePort(t *testing.T) {
	tests := []struct {
		name    string
		want    int
		wantErr bool
	}{
		{name: "string", want: 200, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRandomAvailablePort()
			if (err != nil) != tt.wantErr {
				t.Errorf("getRandomAvailablePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getRandomAvailablePort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	type args struct {
		clientID     string
		clientSecret string
		scopes       string
		authURL      string
		tokenURL     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "string", args: args{clientID: "string", clientSecret: "string", scopes: "string", authURL: "https://localhost:8080/oauth", tokenURL: "https://localhost:8080/oauth"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := LoginOptions{
				ClientID:     tt.args.clientID,
				ClientSecret: tt.args.clientSecret,
				Scopes:       tt.args.scopes,
				AuthURL:      tt.args.authURL,
				TokenURL:     tt.args.tokenURL,
			}
			_, err := Login(options)
			if err != nil {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
