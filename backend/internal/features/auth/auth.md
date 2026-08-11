# Authentication

This will have everything related to authentication and authorization,  
such as register, login, verify email/phone, verify 2fa, etc..

the reason we use channel and purpose is to separate the why and where,  
channel represents the where and purpose represents the why.  
there are soft rules against where each can be passed so it is the developer's duty to make sure they do not pass a unsupported channel where it is not supported, the comments for the channel already make this clear if a certain channel is valid for some certain task only

current todos are:
- fix the database errors, as db is not implemented as of now i have no idea how to handle those errors in the service layer so will have to do that once i implement the db layer

- in register there are 2 calls to the db, one to check if the email/phone exists already and one to check if the username exists already. If possible combine them into a call

