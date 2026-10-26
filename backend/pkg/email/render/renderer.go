package render

import (
	"embed"
	"fmt"
	"html/template"
	"strings"
)

const (
	templateDir = "templates"
	welcomePath = "welcome.html"
	otpPath     = "otp.html"
)

type HTML = string

type Subject int

// enum subjects for emails, the exact string is found in the constructor for Renderer [NewRenderer]
const (
	SubWelcome Subject = iota
	SubVerifyEmail
	Sub2FA
	SubResetPassword
)

const (
	otpCon2FA             = "Two-factor authorization code"
	otpSubCon2FA          = "Use the Two-factor authorization code below to log into your account."
	otpConRegistration    = "Verify your account"
	otpSubConRegistration = "Use the verification code below to continue with your account."
	otpConResetPass       = "Reset password"
	otpSubConResetPass    = "Use this one time authorization code below to reset your password on your account."
)

//go:embed templates/*.html
var templateFS embed.FS

type Renderer struct {
	templates *template.Template
	data      data
	subjects  subjects
}

type subjects [4]string // contains the email subject strings, will increase the array length once there are more subjects

// reusable data that is decided at runtime
type data struct {
	AppName         string
	AppLogoUrl      string
	AppHomePageLink string
}

// Returns a new email renderer using embedded html files, if parsing file system fails then error is returned
func New(appName, appLogoUrl, appHomePageLink string) (*Renderer, error) {
	t, err := template.ParseFS(templateFS, templateDir+"/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{
		templates: t,
		data: data{
			AppName:         appName,
			AppLogoUrl:      appLogoUrl,
			AppHomePageLink: appHomePageLink,
		},
		subjects: subjects{ // the [Subject] enums are the index for this slice
			"Welcome to " + appName,
			"Login verification code for " + appName,
			"Verify your email for " + appName,
			"Reset your password for " + appName,
		},
	}, nil
}

// returns a subject string using the enum type [Subject]
//
// will panic if a invalid subject enum is used
func (r *Renderer) Subject(s Subject) string {
	return r.subjects[s]
}

func (r *Renderer) render(path string, data any) HTML {
	var buf strings.Builder
	buf.Grow(5000) // enough for a html, change when adding bigger html files
	err := r.templates.ExecuteTemplate(&buf, path, data)
	if err != nil {
		panic(fmt.Errorf("template panic while email rendering, %w", err))
	}
	return buf.String()
}

// ? ---- email rendering related methods ----

type WelcomeData struct {
	AppName         string
	AppLogoUrl      string
	Username        string
	AppHomePageLink string
}

// renders a html using the welcome template
func (r *Renderer) Welcome(username string) HTML {
	return r.render(welcomePath, WelcomeData{r.data.AppName, r.data.AppLogoUrl, username, r.data.AppHomePageLink})
}

type otpData struct {
	AppName    string
	AppLogoUrl string
	Content    string
	SubContent string
	Otp        string
	ExpiryTime string
}

func (r *Renderer) Otp(otp, expiryTime string) *wrappedOtpData {
	return &wrappedOtpData{
		r: r,
		o: otpData{
			AppName:    r.data.AppName,
			AppLogoUrl: r.data.AppLogoUrl,
			Otp:        otp,
			ExpiryTime: expiryTime,
		},
	}
}

type wrappedOtpData struct {
	r *Renderer
	o otpData
}

func (w *wrappedOtpData) Registration() HTML {
	w.o.Content = otpConRegistration
	w.o.SubContent = otpSubConRegistration
	return w.r.render(otpPath, w.o)
}

func (w *wrappedOtpData) TwoFA() HTML {
	w.o.Content = otpCon2FA
	w.o.SubContent = otpSubCon2FA
	return w.r.render(otpPath, w.o)
}

func (w *wrappedOtpData) ResetPass() HTML {
	w.o.Content = otpConResetPass
	w.o.SubContent = otpSubConResetPass
	return w.r.render(otpPath, w.o)
}
