package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/EmotionlessDev/sso/internal/domain/models"
	"github.com/EmotionlessDev/sso/internal/jwt"
	"github.com/EmotionlessDev/sso/internal/storage"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrorInvalidCredentials = errors.New("invalid credentials")
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	TokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (userID int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

// New creates a new Auth service
func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	TokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		TokenTTL:     TokenTTL,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int64,
) (token string, err error) {
	const op = "auth.Auth.Login"
	log := a.log.With(slog.String("op", op))
	log.Info("logging in user", slog.String("email", email))

	user, err := a.userProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidCredentials) {
			return "", status.Error(codes.InvalidArgument, "invalid credentials")
		}
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", status.Error(codes.NotFound, "user not found")
		}
		return "", status.Error(codes.Internal, err.Error())
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slog.String("email", email))
		return "", status.Error(codes.InvalidArgument, "invalid credentials")
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			return "", status.Error(codes.NotFound, "app not found")
		}
		return "", status.Error(codes.Internal, err.Error())
	}

	token, err = jwt.NewToken(user, app, a.TokenTTL)
	if err != nil {
		return "", status.Error(codes.Internal, err.Error())
	}

	log.Info("user logged in", slog.String("email", email))
	return token, nil
}

func (a *Auth) RegisterNewUser(
	ctx context.Context,
	email string,
	password string,
) (userID int64, err error) {
	const op = "auth.Auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op))
	log.Info("registering new user", slog.String("email", email))

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := a.userSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		return 0, fmt.Errorf("failed to save user: %w", err)
	}

	log.Info("user registered", slog.Int64("userID", id))
	return id, nil
}

func (a *Auth) IsAdmin(
	ctx context.Context,
	userID int64,
) (isAdmin bool, err error) {
	const op = "auth.Auth.IsAdmin"
	log := a.log.With(slog.String("op", op))
	log.Info("check for admin privilegies")

	ok, err := a.userProvider.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check for admin privilegies: %w", err)
	}

	log.Info("admin privilegies checked")
	return ok, nil
}

