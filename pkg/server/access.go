package server

import (
	"context"
	"errors"
	accesspb "github.com/100mslive/access/pkg/internal"
	"github.com/100mslive/packages/log"
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"os"
)

func (server *Server) getAdapter() (*gormadapter.Adapter, error) {
	username := os.Getenv("CASBIN_DATABASE_USER")
	password := os.Getenv("CASBIN_DATABASE_PASSWORD")
	host := os.Getenv("CASBIN_DATABASE_HOST")
	datasource := username + ":" + password + "@tcp(" + host + ")/"
	return gormadapter.NewAdapter("mysql", datasource) // Your driver and data source.

}

// AuthorizeToken ...
func (server *Server) AuthorizeToken(ctx context.Context, req *accesspb.AuthorizeTokenRequest) (*accesspb.AuthorizeReply, error) {
	if !server.connected {
		log.Errorf(ErrServerNotConnected.Error())
		return nil, ErrServerNotConnected
	}
	// setup casbin auth rules
	a, err := server.getAdapter()
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}
	e, err := casbin.NewEnforcer("cmd/pkg/casbin/auth_model.conf", a)
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	err = e.LoadPolicy()
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	token := req.GetToken()
	if token == "" {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, errors.New("token field is required")
	}

	resource := req.GetResource()
	if resource == "" {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, errors.New("resource field is required")
	}

	action := req.GetAction()
	if action == "" {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, errors.New("action field is required")
	}

	res, err := server.auth.ValidateManagementToken(ctx, token)
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	customer := res.CustomerID
	user := res.UserID

	subject := user + "_" + customer

	logger.Println(subject)

	allowed, err := e.Enforce(subject, resource, action)
	// ok, reason, err := e.EnforceEx(subject, "data", "read")
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	// logger.Println(ok, reason)

	return &accesspb.AuthorizeReply{
		Authorized: allowed,
	}, nil
}

// AuthorizeToken ...
func (server *Server) Authorize(ctx context.Context, req *accesspb.AuthorizeRequest) (*accesspb.AuthorizeReply, error) {
	if !server.connected {
		log.Errorf(ErrServerNotConnected.Error())
		return nil, ErrServerNotConnected
	}
	// setup casbin auth rules
	a, err := server.getAdapter()
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}
	e, err := casbin.NewEnforcer("cmd/auth_model.conf", a)
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	err = e.LoadPolicy()
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	subject := req.GetSubject()
	if subject == "" {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, errors.New("token field is required")
	}

	resource := req.GetResource()
	if resource == "" {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, errors.New("resource field is required")
	}

	action := req.GetAction()
	if action == "" {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, errors.New("action field is required")
	}

	logger.Println(subject)

	allowed, err := e.Enforce(subject, resource, action)
	// ok, reason, err := e.EnforceEx(subject, "data", "read")
	if err != nil {
		return &accesspb.AuthorizeReply{
			Authorized: false,
		}, err
	}

	return &accesspb.AuthorizeReply{
		Authorized: allowed,
	}, nil
}
