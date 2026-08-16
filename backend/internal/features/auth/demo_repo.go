package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type DemoRepository struct {
	users         []*user
	user_sessions []*userSession
}

func NewDemoRepository() *DemoRepository {
	return &DemoRepository{
		make([]*user, 0),
		make([]*userSession, 0),
	}
}

// todos
func (dr *DemoRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*user, error) {
	for _, u := range dr.users {
		if u==nil{
			continue
		}
		if u.ID == userID {
			return u, nil
		}
	}
	return nil, errors.New("not found")
}
func (dr *DemoRepository) FindUserByChannel(ctx context.Context, channel authChannel, target string) (*user, error) {
	var t string
	for _, u := range dr.users {
		if u==nil{
			continue
		}
		switch channel {
		case channelEmail:
			t = u.Email
		case channelPhone:
			t = u.Phone
		case channelUsername:
			t = u.Username
		}
		if t==target{
			return u,nil
		}
	}
	return nil, errors.New("not found")
}
func (dr *DemoRepository) CreateUser(ctx context.Context, user *user) error {
	dr.users = append(dr.users, user)
	return nil
}
func (dr *DemoRepository) UpdateUser(ctx context.Context, userID uuid.UUID, updatedUser *user) error {
	for i,u:=range dr.users{
		if u==nil{
			continue
		}
		if u.ID==userID{
			dr.users[i]=updatedUser
			return nil
		}
	}
	return errors.New("not found")
}

// user_sessions table{}
func (dr *DemoRepository) FindSessionByToken(ctx context.Context, tokenHash string) (*userSession, error) {
	for _,u:=range dr.user_sessions{
		if u==nil{
			continue
		}
		if u.TokenHash==tokenHash{
			return u,nil
		}
	}
	return nil, errors.New("not found")
}
func (dr *DemoRepository) CreateSession(ctx context.Context, session *userSession) error {
	dr.user_sessions = append(dr.user_sessions, session)
	return nil
}
func (dr *DemoRepository) UpdateSessionToken(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	for _,u:=range dr.user_sessions{
		if u==nil{
			continue
		}
		if u.ID==id{
			u.TokenHash=tokenHash
			u.ExpiresAt=expiresAt
			return nil
		}
	}
	return errors.New("not found")
}
func (dr *DemoRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	for i,u:=range dr.user_sessions{
		if u==nil{
			continue
		}
		if u.ID==id{
			dr.user_sessions=append(dr.user_sessions[:i],dr.user_sessions[i+1:]...)	
			return nil
		}
	}
	return errors.New("not found")
}
func (dr *DemoRepository) DeleteSessionByJwtID(ctx context.Context, jwtID uuid.UUID) error {
	for i,u:=range dr.user_sessions{
		if u==nil{
			continue
		}
		if u.JwtID==jwtID{
			dr.user_sessions=append(dr.user_sessions[:i],dr.user_sessions[i+1:]...)	
			return nil
		}
	}
	return errors.New("not found")
}
