<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import LoginScreen from './acerola-login-screen.svelte';

  const idle = { isSubmitting: false, error: null, fieldErrors: {} };
  const data = {
    isMockMode: false,
    signIn: { values: { email: '', password: '' } },
    firstAccess: { values: { name: '', email: '', password: '', passwordConfirmation: '' } },
  };

  const { Story } = defineMeta({
    title: 'Features/Login/AcerolaLoginScreen',
    component: LoginScreen,
    args: {
      state: { signIn: idle, firstAccess: idle },
      actions: {
        signIn: { onChange: fn(), onSubmit: fn() },
        firstAccess: { onChange: fn(), onSubmit: fn() },
      },
    },
  });
</script>

<Story name="Default" args={{ data }} />

<Story
  name="MockMode"
  args={{
    data: {
      ...data,
      isMockMode: true,
      signIn: { values: { email: 'voce@exemplo.com', password: 'artifact-docs' } },
    },
  }}
/>

<!-- Caso limite: a conta entrou na Neon mas não está na lista permitida. -->
<Story
  name="AccessDenied"
  args={{
    data,
    state: {
      signIn: {
        ...idle,
        error: 'Esta conta não tem acesso a este hub. Entre com a conta cadastrada.',
      },
      firstAccess: idle,
    },
  }}
/>
