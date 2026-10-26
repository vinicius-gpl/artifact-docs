package store

import "context"

// HasOwner responde se alguém da lista permitida já criou a conta no Neon Auth. Enquanto
// não, a tela de login oferece o "Primeiro acesso" (GET /api/auth/setup, docs/auth.md).
//
// Lista vazia é "ninguém pode entrar": nem adianta consultar, e o primeiro acesso aparece
// (a conta criada vai receber 403 até alguém configurar AUTH_ALLOWED_EMAILS).
func (store *Store) HasOwner(ctx context.Context, allowedEmails []string) (bool, error) {
	if len(allowedEmails) == 0 {
		return false, nil
	}

	var exists bool
	// "user" entre aspas: é palavra reservada do Postgres, e é o nome que o Neon Auth usa.
	err := store.pool.QueryRow(ctx,
		`select exists (select 1 from neon_auth."user" where lower(email) = any($1))`,
		allowedEmails,
	).Scan(&exists)

	return exists, err
}
