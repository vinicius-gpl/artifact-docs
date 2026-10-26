import { z } from 'zod';

/**
 * Quem está usando o hub — a FORMA do que a API devolve em `GET /api/auth/me`.
 *
 * Sem papel (`role`): o hub é de uma pessoa só. Quem entra é decidido pela lista de e-mails
 * permitidos na API, não por um papel guardado aqui.
 *
 * O `id` é o id da pessoa no Neon Auth. O lado Go mora em `artifact-docs/src-go/` — mudou aqui,
 * muda lá.
 */
export const sessionUserSchema = z.object({
  id: z.string().min(1),
  email: z.string().email(),
  name: z.string().min(1),
  image: z.string().nullable(),
});

export type SessionUser = z.infer<typeof sessionUserSchema>;
