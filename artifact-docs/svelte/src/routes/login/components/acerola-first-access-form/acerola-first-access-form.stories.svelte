<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import FirstAccessForm from './acerola-first-access-form.svelte';

  const values = { name: '', email: '', password: '', passwordConfirmation: '' };
  const idle = { isSubmitting: false, error: null, fieldErrors: {} };

  const { Story } = defineMeta({
    title: 'Features/Login/AcerolaFirstAccessForm',
    component: FirstAccessForm,
    args: { actions: { onChange: fn(), onSubmit: fn() } },
  });
</script>

<Story name="Default" args={{ data: { values }, state: idle }} />

<Story
  name="FieldErrors"
  args={{
    data: { values },
    state: {
      ...idle,
      fieldErrors: {
        name: 'Informe o nome',
        password: 'A senha precisa de pelo menos 8 caracteres',
        passwordConfirmation: 'As senhas não são iguais',
      },
    },
  }}
/>

<Story name="Submitting" args={{ data: { values }, state: { ...idle, isSubmitting: true } }} />

<!-- Caso limite: a conta já foi criada por outra aba. -->
<Story
  name="AlreadyExists"
  args={{
    data: { values },
    state: { ...idle, error: 'A conta do hub já existe. Entre com e-mail e senha.' },
  }}
/>
