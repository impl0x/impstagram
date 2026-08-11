# Impstagram - Social media App
Yes its a instagram clone.
stack-
frontend: flutter (dart)
backend: go, framework: mo
db: postgresql


## Details:
Register:  
- email using resend and custom domain: done setting up api
- phone number using telegram

Json response schema:
{
  "code": "...",
  "message": "...",
  "data": {}
}
with data omitted on errors
and code being a string, such as TWO_FACTOR_REQUIRED, INVALID_TWO_FACTOR_CODE, etc.

## Codebase
to understand the codebase, in most functions every section is labelled with a comment at the top of the section, reading that will explain what that part of the code is doing 