<script lang="ts">
  import { page } from '$app/state';

  import ResetPasswordScreen from './components/acerola-reset-password-screen/acerola-reset-password-screen.svelte';
  import { useResetPasswordModel } from '$lib/hooks/use-reset-password/use-reset-password.svelte';

  /* O token do link vem do endereço. Fora do grupo `(app)`, como o login. */
  const model = useResetPasswordModel({ token: page.url.searchParams.get('token') ?? '' });
</script>

<svelte:head>
  <title>Crie sua senha</title>
</svelte:head>

<ResetPasswordScreen data={model.data} state={model.state} actions={model.actions} />
