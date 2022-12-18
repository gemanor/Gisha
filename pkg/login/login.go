package gisha

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sync"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

var oauthConfig *oauth2.Config
var wg sync.WaitGroup

func getRandomAvailablePort() (int, error) {
	// Bind to port 0 to choose a random available port
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()

	// Return the port number that the listener is using
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func saveRefreshToken(refreshToken string) error {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// Save the refresh token in the keyring using the current user's username
	return keyring.Set("my-app-name", currentUser.Username, refreshToken)
}

func readRefreshToken() (string, error) {
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Read the refresh token from the keyring using the current user's username
	return keyring.Get("my-app-name", currentUser.Username)
}

// Login starts the OAuth flow and logs the user in to their account on the third-party service.
func Login(clientID string, clientSecret string, scopes []string, authURL string, tokenURL string) error {
	// generate random available port number
	port, err := getRandomAvailablePort()
	if err != nil {
		log.Fatalf("Failed to get random available port: %v", err)
	}

	// Set the OAuth configuration for the third-party service
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d/oauth", port),
		Scopes:      scopes,
	}

	// Check for a refresh token in the keyring
	refreshToken, _ := readRefreshToken()
	if refreshToken != "" {
		Refresh(refreshToken)
	}

	// Create a new HTTP server
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	// Register the OAuth callback handler
	http.HandleFunc("/oauth", oauthCallbackHandler)
	wg.Add(1)

	// Start the HTTP server
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// Open the third-party service's OAuth login page in the user's web browser
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOnline)
	if err := openURL(url); err != nil {
		log.Fatalf("Failed to open URL in web browser: %v", err)
	}

	// Wait for the OAuth callback
	wg.Wait()
	httpServer.Shutdown(
		context.Background(),
	)
	return nil
}

// Refresh refreshes the user's access token when it expires.
func Refresh(refreshToken string) (*oauth2.Token, error) {
	// Exchange the refresh token for a new access token
	token, err := oauthConfig.Exchange(context.Background(), refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %v", err)
	}

	// Return the new access token
	return token, nil
}

// oauthCallbackHandler is the HTTP handler for the OAuth callback
func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Get the authorization code or access token from the request
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		wg.Done()
		return
	}

	// Use the authorization code to get an access token
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		wg.Done()
		return
	}

	// Save the refresh token in the keyring
	if err := saveRefreshToken(token.RefreshToken); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		wg.Done()
		return
	}

	// Print the access token to the user
	fmt.Fprintln(os.Stdout, "Token:", token.AccessToken)
	wg.Done()
}

// openURL opens the specified URL in the user's web browser
func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	return exec.Command(cmd, args...).Start()
}
