package gisha

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
	"sync"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

type LoginOptions struct {
	AppID        string
	ClientID     string
	ClientSecret string
	Scopes       string
	AuthURL      string
	TokenURL     string
}

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

func saveRefreshToken(refreshToken string, appID string) error {
	if appID == "" || refreshToken == "" {
		return nil
	}
	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	// Save the refresh token in the keyring using the current user's username
	return keyring.Set(appID, currentUser.Username, refreshToken)
}

func readRefreshToken(appID string) (string, error) {
	if appID == "" {
		return "", nil
	}

	// Get the current user
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}

	// Read the refresh token from the keyring using the current user's username
	log.Println("Reading refresh token from keyring")
	return keyring.Get(appID, currentUser.Username)
}

// Login starts the OAuth flow and logs the user in to their account on the third-party service.
func Login(options LoginOptions) (*oauth2.Token, error) {
	// generate random available port number
	port, err := getRandomAvailablePort()
	if err != nil {
		log.Fatalf("Failed to get random available port: %v", err)
		return nil, err
	}

	// Set the OAuth configuration for the third-party service
	oauthConfig := &oauth2.Config{
		ClientID:     options.ClientID,
		ClientSecret: options.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  options.AuthURL,
			TokenURL: options.TokenURL,
		},
		RedirectURL: fmt.Sprintf("http://localhost:%d/oauth", port),
		Scopes:      strings.Split(options.Scopes, ","),
	}

	// Check for a refresh token in the keyring
	refreshToken, err := readRefreshToken(options.AppID)
	if err != nil && err != keyring.ErrNotFound {
		log.Fatalf("Failed to read refresh token from keyring: %v", err)
		return nil, err
	}
	if refreshToken != "" {
		return refresh(refreshToken, oauthConfig)
	}

	return oauthFlow(port, options.AppID, oauthConfig)
}

func oauthFlow(port int, appID string, oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	// Create an HTTP server to listen for the OAuth callback
	wg := &sync.WaitGroup{}
	token := oauth2.Token{}
	err := error(nil)
	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	wg.Add(1)
	http.HandleFunc("/oauth", oauthCallbackHandler(wg, appID, oauthConfig, &token, &err))

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v. process exists", err)
			wg.Done()
		}
	}()

	// Open the third-party service's OAuth login page in the user's web browser
	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	if err := openURL(url); err != nil {
		log.Fatalf("Failed to open URL in web browser: %v", err)
		return nil, err
	}

	// Wait for the OAuth callback
	wg.Wait()
	httpServer.Shutdown(context.Background())
	return &token, err
}

// Refresh refreshes the user's access token when it expires.
func refresh(refreshToken string, oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	// Exchange the refresh token for a new access token
	token, err := oauthConfig.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	}).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %v", err)
	}

	// Return the new access token
	return token, nil
}

// oauthCallbackHandler is the HTTP handler for the OAuth callback
func oauthCallbackHandler(wg *sync.WaitGroup, appID string, oauthConfig *oauth2.Config, token *oauth2.Token, err *error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization code or access token from the request
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Authorization code not found", http.StatusBadRequest)
			(*err) = fmt.Errorf("authorization code not found")
			wg.Done()
			return
		}

		// Use the authorization code to get an access token
		response, resErr := oauthConfig.Exchange(context.Background(), code)
		if resErr != nil {
			http.Error(w, resErr.Error(), http.StatusInternalServerError)
			(*err) = resErr
			wg.Done()
			return
		}

		// Save the refresh token in the keyring
		if resErr = saveRefreshToken(response.RefreshToken, appID); resErr != nil {
			http.Error(w, resErr.Error(), http.StatusInternalServerError)
			(*err) = resErr
			wg.Done()
			return
		}

		// Store the token, and signal that the OAuth flow is complete
		(*token) = *response
		wg.Done()
	}
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
