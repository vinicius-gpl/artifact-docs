<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import ResetPasswordScreen from './acerola-reset-password-screen.svelte';

  const values = { password: '', passwordConfirmation: '' };
  const idle = { isSubmitting: false, isLinkValid: true, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Features/ResetPassword/AcerolaResetPasswordScreen',
    component: ResetPasswordScreen,
    args: { data: { values }, state: idle, actions: { onChange: fn(), onSubmit: fn() } },
  });
</script>

<Story name="Default" />

<Story
  name="FieldErrors"
  args={{ state: { ...idle, fieldErrors: { passwordConfirmation: 'As senhas não são iguais' } } }}
/>

<Story name="Submitting" args={{ state: { ...idle, isSubmitting: true } }} />

<!-- Caso limite: link vencido ou aberto sem o token. -->
<Story name="InvalidLink" args={{ state: { ...idle, isLinkValid: false } }} />

<Story
  name="Error"
  args={{ state: { ...idle, error: 'Este link não vale mais. Peça um novo no primeiro acesso.' } }}
/>
