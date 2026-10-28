import { z } from 'zod';

/**
 * O CONTRATO do formulário de login: e-mail e senha, nada mais.
 *
 * Quem recebe as credenciais é o **Neon Auth**, direto do navegador — não a nossa API. No modo
 * mock, quem confere é o `mock-server`, com o usuário de teste.
 *
 * Sem `min` de tamanho na senha: a política de senha é do Neon Auth, e repeti-la aqui criaria
 * duas regras para a mesma coisa. Senha curta ou errada volta como "e-mail ou senha
 * incorretos" — a mesma mensagem nos dois casos, para não contar que o e-mail existe.
 */
export const loginRequestSchema = z.object({
  email: z
    .string({ required_error: 'Informe o e-mail' })
    .trim()
    .min(1, 'Informe o e-mail')
    .email('Informe um e-mail válido'),
  password: z.string({ required_error: 'Informe a senha' }).min(1, 'Informe a senha'),
});

export type LoginInput = z.input<typeof loginRequestSchema>;

export const INVALID_CREDENTIALS = 'E-mail ou senha incorretos.';

/** O tamanho mínimo de senha do Neon Auth. Repetido aqui para o erro aparecer ANTES do envio. */
export const PASSWORD_MIN_LENGTH = 8;

/**
 * O CONTRATO do PRIMEIRO ACESSO (e do "esqueci a senha"): só o e-mail, para onde o Neon Auth
 * manda o link de criar a senha.
 *
 * A conta nasce no painel da Neon, sem senha — é por este link que o dono define a primeira.
 * Não é cadastro: e-mail sem conta não recebe nada, e a tela não conta a diferença.
 */
export const firstAccessSchema = z.object({
  email: z
    .string({ required_error: 'Informe o e-mail' })
    .trim()
    .min(1, 'Informe o e-mail')
    .email('Informe um e-mail válido'),
});

export type FirstAccessInput = z.input<typeof firstAccessSchema>;

/**
 * O CONTRATO da SENHA NOVA, aberta pelo link do e-mail.
 *
 * A confirmação existe porque a pessoa não vê o que digita: sem ela, um dedo errado vira uma
 * senha que ninguém conhece. O erro da confirmação aponta para o campo da confirmação — é nele
 * que a pessoa precisa mexer.
 */
export const newPasswordSchema = z
  .object({
    password: z
      .string({ required_error: 'Informe a senha' })
      .min(PASSWORD_MIN_LENGTH, `A senha precisa de pelo menos ${PASSWORD_MIN_LENGTH} caracteres`),
    passwordConfirmation: z.string({ required_error: 'Repita a senha' }).min(1, 'Repita a senha'),
  })
  .refine((values) => values.password === values.passwordConfirmation, {
    message: 'As senhas não são iguais',
    path: ['passwordConfirmation'],
  });

export type NewPasswordInput = z.input<typeof newPasswordSchema>;

/**
 * `GET /api/auth/setup` — pública. `needsSetup` é verdade enquanto não existe conta com um dos
 * e-mails permitidos.
 */
export const setupStatusSchema = z.object({ needsSetup: z.boolean() });

export type SetupStatus = z.infer<typeof setupStatusSchema>;
