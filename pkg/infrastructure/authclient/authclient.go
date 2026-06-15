// Package authclient is the HTTP client for auth-service's internal endpoints.
// The WhatsApp worker uses it to resolve a sender's phone number to a user.
package authclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/kharchibook/expense-service/constants"
)

// ErrNotRegistered is returned when no account matches the phone (auth-service
// 404). The caller turns this into a "please sign up" reply.
var ErrNotRegistered = errors.New("phone not registered")

// User is the resolved identity returned by auth-service.
type User struct {
	UserID int64  `json:"userId"`
	Name   string `json:"name"`
}

// IClient resolves identities against auth-service.
type IClient interface {
	UserByPhone(ctx context.Context, phone string) (*User, error)
}

type client struct {
	baseURL     string
	internalKey string
	http        *http.Client
}

// New constructs the auth-service client.
func New(baseURL, internalKey string) IClient {
	return &client{
		baseURL:     baseURL,
		internalKey: internalKey,
		http:        &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *client) UserByPhone(ctx context.Context, phone string) (*User, error) {
	endpoint := fmt.Sprintf("%s/v1/internal/users/by-phone?phone=%s",
		c.baseURL, url.QueryEscape(phone))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set(constants.HeaderInternalKey, c.internalKey)

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call auth-service: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	switch res.StatusCode {
	case http.StatusOK:
		var u User
		if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}
		return &u, nil
	case http.StatusNotFound:
		return nil, ErrNotRegistered
	default:
		return nil, fmt.Errorf("auth-service returned %d", res.StatusCode)
	}
}
