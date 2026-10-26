package auth

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed queries/has_owner.sql
var hasOwnerSQL string

// OwnerRepository pergunta ao banco do Neon Auth se o dono já tem conta.
type OwnerRepository struct {
	pool *pgxpool.Pool
}

// NewOwnerRepository lê do pool do Neon.
func NewOwnerRepository(pool *pgxpool.Pool) *OwnerRepository {
	return &OwnerRepository{pool: pool}
}

// HasOwner responde se alguém da lista permitida já criou a conta no Neon Auth. Enquanto
// não, a tela de login oferece o "Primeiro acesso".
//
// Lista vazia é "ninguém pode entrar": nem adianta consultar, e o primeiro acesso aparece
// (a conta criada vai receber 403 até alguém configurar AUTH_ALLOWED_EMAILS).
func (repository *OwnerRepository) HasOwner(ctx context.Context, allowedEmails []string) (bool, error) {
	if len(allowedEmails) == 0 {
		return false, nil
	}

	var exists bool
	err := repository.pool.QueryRow(ctx, hasOwnerSQL, allowedEmails).Scan(&exists)

	return exists, err
}
