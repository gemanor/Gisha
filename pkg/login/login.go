package gisha

import (
	"context"
	"errors"
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
	"golang.org/x/oauth2/endpoints"
)

var successfulHTML = `<html><head>
<style>
  body {
	background-color: #f1f1f1;
  }
  .message {
	margin: auto;
	width: 50%;
	border: 3px solid #333;
	padding: 10px;
	text-align: center;
	border-radius: 5px;
	box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
	transition: 0.3s;
	background-color: #fff;
  }
  .message:hover {
	box-shadow: 0 8px 16px 0 rgba(0,0,0,0.2);
  }
  h1 {
	color: #333;
	font-size: 32px;
  }
  p {
	font-size: 18px;
  }
  .icon {
	display: inline-block;
	width: 60px;
	height: 60px;
	background-color: #4caf50;
	border-radius: 50%;
	color: #fff;
	line-height: 60px;
	font-size: 36px;
	box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
	transition: 0.3s;
  }
  .icon:hover {
	box-shadow: 0 8px 16px 0 rgba(0,0,0,0.2);
  }
  .icon::before {
	content: "";
	display: block;
	position: relative;
	top: 18px;
	left: 10px;
	width: 30px;
	height: 10px;
	border-left: 10px solid #fff;
	border-bottom: 10px solid #fff;
	transform: rotate(-45deg);
  }
</style>
</head>
<body>
<div class="message"><div class="icon"></div>
  
  <h1>Authentication Successful!</h1>
  <p>You can now close this page.</p>
</div>

</body>
</html>`

var failureHTML = `<html><head>
<style>
  body {
	background-color: #f1f1f1;
  }
  .message {
	margin: auto;
	width: 50%;
	border: 3px solid #333;
	padding: 10px;
	text-align: center;
	border-radius: 5px;
	box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
	transition: 0.3s;
	background-color: #fff;
  }
  .message:hover {
	box-shadow: 0 8px 16px 0 rgba(0,0,0,0.2);
  }
  h1 {
	color: #333;
	font-size: 32px;
  }
  p {
	font-size: 18px;
  }
  .icon {
display: inline-block;
width: 60px;
height: 60px;
background-color: #cc0000;
border-radius: 50%;
color: #fff;
line-height: 60px;
font-size: 36px;
box-shadow: 0 4px 8px 0 rgba(0,0,0,0.2);
transition: 0.3s;
}
.icon:hover {
box-shadow: 0 8px 16px 0 rgba(0,0,0,0.2);
}
.icon::before,
.icon::after {
	content: "";
    display: block;
    position: absolute;
    top: 49px;
    left: 0;
    right: 0;
    margin: auto;
    width: 30px;
    height: 5px;
    background-color: #fff;
}
.icon::before {
transform: rotate(45deg);
}
.icon::after {
transform: rotate(-45deg);
}
</style>
</head>
<body>
<div class="message"><div class="icon"></div>
  
  <h1>Authentication Failed</h1>
  <p>There was an error during the authentication process. Please try again.</p>
  
</div>

</body></html>`

type Provider string

const (
	Google           Provider = "google"
	Microsoft        Provider = "microsoft"
	Facebook         Provider = "facebook"
	LinkedIn         Provider = "linkedin"
	GitHub           Provider = "github"
	Spotify          Provider = "spotify"
	Slack            Provider = "slack"
	Yahoo            Provider = "yahoo"
	Fitbit           Provider = "fitbit"
	Strava           Provider = "strava"
	Uber             Provider = "uber"
	Amazon           Provider = "amazon"
	Battlenet        Provider = "battlenet"
	Cern             Provider = "cern"
	FourSquare       Provider = "FourSquare"
	GitLab           Provider = "gitlab"
	Heroku           Provider = "heroku"
	Instagram        Provider = "instagram"
	Mailchimp        Provider = "mailchimp"
	Mailru           Provider = "mailru"
	HipChat          Provider = "hipchat"
	MediaMath        Provider = "mediamath"
	NokiaHealth      Provider = "nokiahealth"
	Odnoklassniki    Provider = "odnoklassniki"
	PayPal           Provider = "paypal"
	Bitbucket        Provider = "bitbucket"
	MediaMathSandbox Provider = "mediamathsandbox"
	StackOverflow    Provider = "stackoverflow"
	Twitch           Provider = "twitch"
	VK               Provider = "vk"
	Yandex           Provider = "yandex"
	Zoom             Provider = "zoom"
)

// String is used both by fmt.Print and by Cobra in help text
func (e *Provider) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (e *Provider) Set(v string) error {
	switch v {
	case "google", "microsoft", "facebook", "linkedin", "github", "spotify", "slack", "yahoo", "fitbit", "strava", "uber", "amazon", "battlenet", "cern", "FourSquare", "gitlab", "heroku", "instagram", "mailchimp", "mailru", "hipchat", "mediamath", "nokiahealth", "odnoklassniki", "paypal", "bitbucket", "mediamathsandbox", "stackoverflow", "twitch", "vk", "yandex", "zoom":
		*e = Provider(v)
		return nil
	default:
		return errors.New(`Must be one of "google", "microsoft", "facebook", "linkedin", "github", "spotify", "slack", "yahoo", "fitbit", "strava", "uber", "amazon", "battlenet", "cern", "FourSquare", "gitlab", "heroku", "instagram", "mailchimp", "mailru", "hipchat", "mediamath", "nokiahealth", "odnoklassniki", "paypal", "bitbucket", "mediamathsandbox", "stackoverflow", "twitch", "vk", "yandex", "zoom"`)
	}
}

// Type is only used in help text
func (e *Provider) Type() string {
	return "provider"
}

var providers = map[string]oauth2.Endpoint{
	"google":           endpoints.Google,
	"microsoft":        endpoints.Microsoft,
	"facebook":         endpoints.Facebook,
	"linkedin":         endpoints.LinkedIn,
	"github":           endpoints.GitHub,
	"spotify":          endpoints.Spotify,
	"slack":            endpoints.Slack,
	"yahoo":            endpoints.Yahoo,
	"fitbit":           endpoints.Fitbit,
	"strava":           endpoints.Strava,
	"uber":             endpoints.Uber,
	"amazon":           endpoints.Amazon,
	"battlenet":        endpoints.Battlenet,
	"cern":             endpoints.Cern,
	"FourSquare":       endpoints.Foursquare,
	"gitlab":           endpoints.GitLab,
	"heroku":           endpoints.Heroku,
	"instagram":        endpoints.Instagram,
	"mailchimp":        endpoints.Mailchimp,
	"mailru":           endpoints.Mailru,
	"hipchat":          endpoints.HipChat,
	"mediamath":        endpoints.MediaMath,
	"nokiahealth":      endpoints.NokiaHealth,
	"odnoklassniki":    endpoints.Odnoklassniki,
	"paypal":           endpoints.PayPal,
	"bitbucket":        endpoints.Bitbucket,
	"mediamathsandbox": endpoints.MediaMathSandbox,
	"stackoverflow":    endpoints.StackOverflow,
	"twitch":           endpoints.Twitch,
	"vk":               endpoints.Vk,
	"yandex":           endpoints.Yandex,
	"zoom":             endpoints.Zoom,
}

type LoginOptions struct {
	AppID        string
	ClientID     string
	ClientSecret string
	Scopes       string
	AuthURL      string
	TokenURL     string
	Provider     Provider
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
		Endpoint:     endpoints.Google,
		RedirectURL:  fmt.Sprintf("http://localhost:%d/oauth", port),
		Scopes:       strings.Split(options.Scopes, ","),
	}

	token, _ := refresh(options.AppID, oauthConfig)
	if token != nil {
		return token, nil
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
func refresh(appID string, oauthConfig *oauth2.Config) (*oauth2.Token, error) {
	// Check for a refresh token in the keyring
	refreshToken, err := readRefreshToken(appID)
	if err != nil && err != keyring.ErrNotFound {
		log.Fatalf("Failed to read refresh token from keyring: %v", err)
		return nil, err
	}
	// Exchange the refresh token for a new access token
	token, _ := oauthConfig.TokenSource(context.Background(), &oauth2.Token{
		RefreshToken: refreshToken,
	}).Token()

	// Return the new access token
	return token, nil
}

// oauthCallbackHandler is the HTTP handler for the OAuth callback
func oauthCallbackHandler(wg *sync.WaitGroup, appID string, oauthConfig *oauth2.Config, token *oauth2.Token, err *error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the authorization code or access token from the request
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, failureHTML, http.StatusBadRequest)
			(*err) = fmt.Errorf("authorization code not found")
			wg.Done()
			return
		}

		// Use the authorization code to get an access token
		response, resErr := oauthConfig.Exchange(context.Background(), code)
		if resErr != nil {
			http.Error(w, failureHTML, http.StatusBadRequest)
			(*err) = resErr
			wg.Done()
			return
		}

		// Save the refresh token in the keyring
		if resErr = saveRefreshToken(response.RefreshToken, appID); resErr != nil {
			http.Error(w, failureHTML, http.StatusBadRequest)
			(*err) = resErr
			wg.Done()
			return
		}

		// Store the token, and signal that the OAuth flow is complete
		(*token) = *response
		fmt.Fprint(w, successfulHTML)
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
