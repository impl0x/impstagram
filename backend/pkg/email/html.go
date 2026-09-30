package email

import (
	"backend/internal/config"
	"strings"
)

// yes i know this is messy and repetitive. and unsafe
// these html's are ai generated, i don't know html nor do i have the energy to learn html and format everything dynamically as of now
// i have only 2 email cases so this is enough for me as of now

type HTML struct {
	Html            string
	MainPlaceholder string
	Placeholders    []string
}

func (h HTML) Format(mainReplacement string) string {
	finalSlice := make([]string, len(h.Placeholders)+2)
	finalSlice = append(finalSlice, h.Placeholders...)
	finalSlice = append(finalSlice, h.MainPlaceholder, mainReplacement)
	return strings.NewReplacer(finalSlice...).Replace(h.Html)
}

var HtmlWelcome = HTML{`
<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Welcome to Impstagram</title>
</head>

<body style="margin:0; padding:0; background-color:#0f0f0f; font-family:Arial, Helvetica, sans-serif; color:#f5f5f5;">

  <table width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#0f0f0f;">
    <tr>
      <td align="center" style="padding:50px 20px;">

        <!-- Main Card -->
        <table width="100%" cellpadding="0" cellspacing="0" border="0"
          style="max-width:560px; background-color:#181818; border:1px solid #2a2a2a; border-radius:14px; overflow:hidden;">

          <!-- Header -->
          <tr>
            <td align="center" style="padding:42px 30px 25px;">

              <!-- Logo -->
              <a href="https://ibb.co/0yq5Vf8C" target="_blank" style="display: inline-block; text-decoration: none;">
                <img src="https://i.ibb.co/b51tRrqs/impstagramlogo.jpg" alt="Logo" width="100"
                  style="display: block; width: 100px; height: auto; border: 0; border-radius: 20px; -webkit-border-radius: 12px; -moz-border-radius: 12px;">
              </a>


              <div style="
                font-size:30px;
                font-weight:700;
                letter-spacing:-1px;
                color:#ffffff;
              ">
                Impstagram
              </div>

            </td>
          </tr>

          <!-- Content -->
          <tr>
            <td style="padding:10px 45px 40px;">

              <h1 style="
                margin:0 0 18px;
                font-size:24px;
                line-height:32px;
                font-weight:600;
                color:#ffffff;
              ">
                Welcome to Impstagram, {{username}}
              </h1>

              <p style="
                margin:0 0 20px;
                font-size:16px;
                line-height:26px;
                color:#b8b8b8;
              ">
                Your account is ready. Impstagram is a place to share
                moments, discover new people, and keep up with the things
                that matter to you.
              </p>

              <p style="
                margin:0 0 30px;
                font-size:16px;
                line-height:26px;
                color:#b8b8b8;
              ">
                Start sharing your world and see what everyone else is up to.
              </p>

              <!-- Button -->
              <table cellpadding="0" cellspacing="0" border="0" align="center">
                <tr>
                  <td style="
					border-radius:8px;
					background:linear-gradient(90deg,#d629c2,#2fb3bf);
					">
                    <a href="https://www.youtube.com/watch?v=dQw4w9WgXcQ" target="_blank" style="
						display:inline-block;
						padding:13px 26px;
						font-size:15px;
						font-weight:600;
						color:#ffffff;
						text-decoration:none;
					">
                      Get started
                    </a>
                  </td>
                </tr>
              </table>

              <!-- Divider -->
          <tr>
            <td style="padding:0 45px;">
              <div style="height:1px; background-color:#2a2a2a;"></div>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td align="center" style="padding:25px 30px 30px;">

              <p style="
                margin:0 0 8px;
                font-size:13px;
                color:#777777;
              ">
                You're receiving this email because an Impstagram
                account was created with this address.
              </p>

              <p style="
                margin:0;
                font-size:12px;
                color:#555555;
              ">
                © 2026 Impstagram
              </p>

            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>

</body>

</html>
`, "{{username}}", nil}

// verification
// content: Verify your account
// sub-content: Use the verification code below to continue with your Impstagram account.
// 2fa
// content: Two-factor authorization code
// sub-content: Use the Two-factor authorization code below to log into your Impstagram account.
var HtmlRegistrationVerificationOTP = HTML{
	OTPBaseHTML, "{{otp}}", []string{
		"{{content}}", "Verify your account",
		"{{sub-content}}", "Use the verification code below to continue with your " + config.ServiceName + " account."},
}
var Html2FAOTP = HTML{
	OTPBaseHTML, "{{otp}}", []string{"{{content}}", "Two-factor authorization code", "{{sub-content}}", "Use the Two-factor authorization code below to log into your " + config.ServiceName + " account."},
}

var HtmlResetPasswordOTP = HTML{
	OTPBaseHTML, "{{otp}}", []string{"{{content}}", "Reset password", "{{sub-content}}", "Use this one time authorization code below to reset your password on your " + config.ServiceName + " account."},
}

var OTPBaseHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Your Impstagram verification code</title>
</head>

<body style="margin:0; padding:0; background-color:#0f0f0f; font-family:Arial, Helvetica, sans-serif; color:#f5f5f5;">

  <table width="100%" cellpadding="0" cellspacing="0" border="0" style="background-color:#0f0f0f;">
    <tr>
      <td align="center" style="padding:50px 20px;">

        <table width="100%" cellpadding="0" cellspacing="0" border="0"
          style="max-width:560px; background-color:#181818; border:1px solid #2a2a2a; border-radius:14px; overflow:hidden;">

          <!-- Header -->
          <tr>
            <td align="center" style="padding:42px 30px 25px;">

              <a href="https://ibb.co/0yq5Vf8C" target="_blank" style="display: inline-block; text-decoration: none;">
                <img src="https://i.ibb.co/b51tRrqs/impstagramlogo.jpg" alt="Logo" width="100"
                  style="display: block; width: 100px; height: auto; border: 0; border-radius: 20px; -webkit-border-radius: 12px; -moz-border-radius: 12px;">
              </a>

              <div style="
                font-size:30px;
                font-weight:700;
                letter-spacing:-1px;
                color:#ffffff;
              ">
                Impstagram
              </div>

            </td>
          </tr>

          <!-- Content -->
          <tr>
            <td style="padding:10px 45px 40px;">

              <h1 style="
                margin:0 0 18px;
                font-size:24px;
                line-height:32px;
                font-weight:600;
                color:#ffffff;
                text-align:center;
              ">
                {{content}}
              </h1>

              <p style="
                margin:0 0 25px;
                font-size:16px;
                line-height:26px;
                color:#b8b8b8;
                text-align:center;
              ">
                {{sub-content}}
              </p>

              <!-- OTP -->
              <table width="100%" cellpadding="0" cellspacing="0" border="0">
                <tr>
                  <td align="center" style="padding:5px 0 28px;">

                    <div style="
                      display:inline-block;
                      padding:16px 28px;
                      background-color:#222222;
                      border:1px solid #333333;
                      border-radius:10px;
                      font-size:32px;
                      line-height:38px;
                      font-weight:700;
                      letter-spacing:8px;
                      color:#ffffff;
                    ">
                      {{otp}}
                    </div>

                  </td>
                </tr>
              </table>

              <p style="
                margin:0;
                font-size:14px;
                line-height:22px;
                color:#888888;
                text-align:center;
              ">
                This code will expire in <strong style="color:#b8b8b8;">10 minutes</strong>.
              </p>

              <p style="
                margin:20px 0 0;
                font-size:14px;
                line-height:22px;
                color:#666666;
                text-align:center;
              ">
                If you didn't request this code, you can safely ignore this email.
              </p>

            </td>
          </tr>

          <!-- Divider -->
          <tr>
            <td style="padding:0 45px;">
              <div style="height:1px; background-color:#2a2a2a;"></div>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td align="center" style="padding:25px 30px 30px;">

              <p style="
                margin:0;
                font-size:12px;
                color:#555555;
              ">
                © 2026 Impstagram
              </p>

            </td>
          </tr>

        </table>

      </td>
    </tr>
  </table>

</body>
</html>
`
