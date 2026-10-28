<script module lang="ts">
  import { defineMeta } from '@storybook/addon-svelte-csf';
  import { fn } from 'storybook/test';

  import FirstAccessScreen from './acerola-first-access-screen.svelte';

  const idle = { isSubmitting: false, isSent: false, error: null };

  const { Story } = defineMeta({
    title: 'Features/FirstAccess/AcerolaFirstAccessScreen',
    component: FirstAccessScreen,
    args: { data: { email: '' }, state: idle, actions: { onChange: fn(), onSubmit: fn() } },
  });
</script>

<Story name="Default" />

<Story
  name="FieldError"
  args={{ data: { email: 'dono' }, state: { ...idle, fieldError: 'Informe um e-mail válido' } }}
/>

<Story
  name="Submitting"
  args={{ data: { email: 'dono@exemplo.com' }, state: { ...idle, isSubmitting: true } }}
/>

<Story name="Sent" args={{ data: { email: 'dono@exemplo.com' }, state: { ...idle, isSent: true } }} />

<!-- Caso limite: o Neon Auth recusou por excesso de tentativas. -->
<Story
  name="Error"
  args={{
    data: { email: 'dono@exemplo.com' },
    state: { ...idle, error: 'Muitas tentativas seguidas. Espere alguns minutos e tente de novo.' },
  }}
/>
