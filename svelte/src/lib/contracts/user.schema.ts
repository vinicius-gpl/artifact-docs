import { z } from 'zod';

/**
 * Identidade — a FORMA de quem está usando o sistema.
 *
 * Quem autentica é o **Neon Auth**: a tela manda e-mail e senha direto para lá, recebe um
 * token assinado e o envia à API a cada requisição. A API (Go) confere a assinatura (JWKS) e
 * devolve quem é a pessoa em `/api/auth/me`. Este arquivo não é o cadastro — é só o que a
 * tela precisa para mostrar quem entrou.
 *
 * Não existe pacote `shared/` aqui: o backend é Go, e o contrato do lado dele mora em
 * `src-go/`. Os dois lados precisam combinar — mudou aqui, muda lá.
 */
export const userRoleSchema = z.enum(['user', 'admin']);

export type UserRole = z.infer<typeof userRoleSchema>;

export const USER_ROLE_LABELS: Record<UserRole, string> = {
  user: 'Usuário',
  admin: 'Administrador',
};

export function userRoleLabel(role: UserRole): string {
  return USER_ROLE_LABELS[role];
}

/** O papel de quem entrou sem papel definido no Neon Auth: o mais restrito que existe. */
export const DEFAULT_USER_ROLE: UserRole = 'user';

/**
 * Quem está usando o sistema agora.
 *
 * O `id` é o id da pessoa no Neon Auth (`neon_auth.user.id`). `SessionUser` nunca carrega
 * senha nem token.
 */
export const sessionUserSchema = z.object({
  id: z.string().min(1),
  email: z.string().email(),
  name: z.string().min(1),
  image: z.string().nullable().optional(),
  role: userRoleSchema,
});

export type SessionUser = z.infer<typeof sessionUserSchema>;
