<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import SignInForm from './acerola-sign-in-form.svelte';

  const values = { email: '', password: '' };
  const idle = { isSubmitting: false, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Features/Login/AcerolaSignInForm',
    component: SignInForm,
    args: { actions: { onChange: fn(), onSubmit: fn() } },
  });
</script>

<Story name="Default" args={{ data: { values }, state: idle }} />

<Story
  name="FieldErrors"
  args={{
    data: { values },
    state: { ...idle, fieldErrors: { email: 'Informe o e-mail', password: 'Informe a senha' } },
  }}
/>

<Story name="Submitting" args={{ data: { values }, state: { ...idle, isSubmitting: true } }} />

<!-- Caso limite: credencial recusada. -->
<Story
  name="Refused"
  args={{ data: { values }, state: { ...idle, error: 'E-mail ou senha incorretos.' } }}
/>
