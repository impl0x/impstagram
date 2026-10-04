# Authentication

This will have everything related to authentication and authorization,  
such as register, login, verify email/phone, verify 2fa, etc..

the reason we use channel and purpose is to separate the why and where,  
channel represents the where and purpose represents the why.  
there are soft rules against where each can be passed so it is the developer's duty to make sure they do not pass a unsupported channel where it is not supported, the comments for the channel already make this clear if a certain channel is valid for some certain task only

## Authorization and tokens

the way authorization and tokens work in this app is that there are 2 tokens, access tokens and refresh tokens.  
both are given to the user at login, access token last for a short period of time, defined in the config/constants.AccessTokenExpiryTime.  
The access token is a jwt with 4 fields, user id, issued at, expires at, and jwt id.  
when the access token expires the user has to send their refresh token and refresh both their tokens and a new set of tokens is returned and the old refresh token is updated and no longer valid.  
refresh tokens last longer, defined in config/constants.RefreshTokenExpiryTime.  
Whenever the user hits the refresh endpoint with a valid unexpired refresh token we update the expiry time to the max again,  
it acts as a sort of a sliding window, so if the expiry period of a refresh token is 30 days the user needs to at least refresh the token once in order to stay logged in.  
Access tokens are used the most for mostly every authorization, and these last for a short period of time because these are stateless, and refresh tokens are stateful opaque tokens.  
When generating a pair of access and refresh tokens on login, we store the refresh tokens in a table called user sessions which contains the user metadata such as ip address, device info, refresh token hash, and of course created at expires at.  
Other than that we also store the the access token jwt's jwt id in the user session. And whenever refreshing the access token we never change the jwt id, it is essentially linked to the user session that was created when the user logged in for the first time.  
this makes it easier for the user to be logged in through different devices.  
when a user hits the log out endpoint the first thing we do is retrieve the authorization jwt access token's id and search it in our user sessions table, if found we remove that user session row and add the jwt id to a in memory blacklist till the jwt expires, this ensures that if the user logs out then tries to access something using the same jwt token which just was successfully logged out then it cant access stuff and gets blocked by the in memory cache. This successfully eliminates a edge case where even after logging out the jwt will be valid till it expires as it is stateless. 

## Request flows for the frontend  

### Registration:  
1. /register with the required details, sends otp to the identifier and returns a reference id
2. /verify with the otp and the reference id, returns access and refresh tokens  
### Login:
1. /login with details, and returns access and refresh tokens if 2fa not present, else sends a otp to the primary 2fa channel and returns a reference id
2. if 2fa /verify with otp and reference id, returns tokens now.  

## todos

current todos are:
- fix the database errors, as db is not implemented as of now i have no idea how to handle those errors in the service layer so will have to do that once i implement the db layer

- in register there are 2 calls to the db, one to check if the email/phone exists already and one to check if the username exists already. If possible combine them into a call

