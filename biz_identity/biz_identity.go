package biz_identity

import (
	"fmt"
	"regexp"
)

var identityIDPattern = regexp.MustCompile(`^[A-Z]+(?:\.[A-Z]+){0,9}$`)

type BizIdentity interface {
	IdentityId() string
	Priority() int
}

type Parser[T BizIdentity] interface {
	Parser(info map[string]string) (T, error)
}

type Validator[T BizIdentity] interface {
	Validate(identity T) error
}

type DefaultValidator[T BizIdentity] struct{}

type ParseFunc[T BizIdentity] func(info map[string]string) (T, error)

func (fn ParseFunc[T]) Parser(info map[string]string) (T, error) {
	return fn(info)
}

type ValidateFunc[T BizIdentity] func(identity T) error

func (fn ValidateFunc[T]) Validate(identity T) error {
	return fn(identity)
}

func (DefaultValidator[T]) Validate(identity T) error {
	return ValidateIdentityID(identity.IdentityId())
}

func ValidateIdentityID(identityID string) error {
	if identityID == "" {
		return fmt.Errorf("identity id is required")
	}
	if !identityIDPattern.MatchString(identityID) {
		return fmt.Errorf("identity id must contain 1 to 10 levels, use '.' as the separator, and only contain uppercase letters")
	}
	return nil
}
