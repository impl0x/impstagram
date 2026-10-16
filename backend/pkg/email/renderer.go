package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"sync"
)

const (
	templateDir = "templates"
	welcomePath = templateDir + "/welcome.html"
	otpPath     = templateDir + "/otp.html"
)

type Subject int

// enum subjects for emails, the exact string is found in the constructor for Renderer [NewRenderer]
const (
	SubVerifyEmail Subject = iota
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

var bufferPool = sync.Pool{ // pool to avoid allocations every time we render a email
	New: func() any {
		return &bytes.Buffer{}
	},
}

//go:embed templates/*.html
var templateFS embed.FS

type Renderer struct {
	templates template.Template
	data      data
	subjects  subjects
}

type subjects [3]string // contains the email subject strings, will increase the array length once there are more subjects

// reusable data that is decided at runtime
type data struct {
	AppName         string
	AppLogoUrl      string
	AppHomePageLink string
}

func NewRenderer(appName, appLogoUrl, appHomePageLink string) (*Renderer, error) {
	t, err := template.ParseFS(templateFS, templateDir+"/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{
		templates: *t,
		data: data{
			AppName:         appName,
			AppLogoUrl:      appLogoUrl,
			AppHomePageLink: appHomePageLink,
		},
		subjects: subjects{
			"Login verification code for " + appName, // the [Subject] enums are the index for this slice
			"Verify your email for " + appName,
			"Reset your password for " + appName,
		},
	}, nil
}

// will panic if a invalid subject enum is used, else a valid subject string is returned
func (r Renderer) Subject(s Subject) string {
	return r.subjects[s]
}

// custom type to avoid conflict with normal string and html string
type HTML string

func (r Renderer) render(path string, data any) HTML {
	buf := bufferPool.Get().(*bytes.Buffer)
	err := r.templates.ExecuteTemplate(buf, path, data)
	if err != nil {
		panic(fmt.Errorf("template panic while email rendering, %w", err))
	}
	return HTML(buf.String())
}

type WelcomeData struct {
	Username        string
	AppHomePageLink string
}

func (r Renderer) Welcome(username, appHomePageLink string) HTML {
	return r.render(welcomePath, WelcomeData{username, appHomePageLink})
}

type otpData struct {
	Content    string
	SubContent string
	Otp        string
	ExpiryTime string
}

type wrappedOtpData struct {
	r *Renderer
	o otpData
}

func (w wrappedOtpData) Registration() HTML {
	w.o.Content = otpConRegistration
	w.o.SubContent = otpSubConRegistration
	return w.r.render(otpPath, w.o)
}

func (w wrappedOtpData) TwoFA() HTML {
	w.o.Content = otpCon2FA
	w.o.SubContent = otpSubCon2FA
	return w.r.render(otpPath, w.o)
}

func (w wrappedOtpData) ResetPass() HTML {
	w.o.Content = otpConResetPass
	w.o.SubContent = otpSubConResetPass
	return w.r.render(otpPath, w.o)
}

func (r Renderer) Otp(otp, expiryTime string) wrappedOtpData {
	return wrappedOtpData{
		r: &r,
		o: otpData{
			Otp:        otp,
			ExpiryTime: expiryTime,
		},
	}
}
