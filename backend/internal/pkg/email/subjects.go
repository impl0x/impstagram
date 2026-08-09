package email

import "backend/internal/config"

const SubjectTwoFa = "Login verification code for " + config.ServiceName

const SubjectVerifyEmail = "Verify your email for " + config.ServiceName