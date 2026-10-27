<script lang="ts" module>
  import { cn } from '$lib/utils/cn';

  /**
   * Campo de ARQUIVO: escolher (ou arrastar) um arquivo, com rótulo em cima e erro embaixo,
   * como todo campo do sistema.
   *
   * O shadcn-svelte não tem campo de arquivo, e o `<input type="file">` nativo não se deixa
   * estilizar — cada navegador desenha o botão e o "nenhum arquivo escolhido" do seu jeito, em
   * inglês. O nativo continua aqui, escondido (`sr-only`) mas focável e anunciado; quem a pessoa
   * vê é o `<label>` dele, que ocupa a linha inteira: clicar em qualquer ponto abre o seletor
   * (Fitts), e soltar um arquivo em cima também vale.
   *
   * O componente não guarda o arquivo: avisa em `onChange` e mostra o nome que voltar em
   * `data.fileName`.
   */
  export type AcerolaFileFieldProps = {
    data: {
      label: string;
      name: string;
      /** O `accept` do input: extensões ou tipos (".html,.zip"). */
      accept?: string;
      /** O nome do arquivo escolhido; `null` mostra o `placeholder`. */
      fileName: string | null;
      placeholder?: string;
      /** Uma linha embaixo do campo, quando não há erro: o que é aceito. */
      hint?: string;
      isRequired?: boolean;
    };
    ui?: { className?: string };
    state?: { error?: string | null; isDisabled?: boolean };
    actions?: { onChange?: (file: File | null) => void; onBlur?: () => void };
  };

  /** O arquivo de um `change` ou de um `drop` — o primeiro, quando vierem vários. */
  export function firstFile(files: FileList | null | undefined): File | null {
    return files?.item(0) ?? null;
  }
</script>

<script lang="ts">
  import Paperclip from '@lucide/svelte/icons/paperclip';

  let { data, ui, state: fieldState, actions }: AcerolaFileFieldProps = $props();

  /* `$state` puramente visual: a linha acende enquanto um arquivo passa por cima dela. */
  let isDragging = $state(false);

  const inputId = `ff-${Math.random().toString(36).slice(2, 9)}`;
  const messageId = `${inputId}-message`;

  const error = $derived(fieldState?.error ?? null);
  const hasError = $derived(Boolean(error));
  const isDisabled = $derived(Boolean(fieldState?.isDisabled));

  function onDrop(event: DragEvent): void {
    event.preventDefault();
    isDragging = false;
    if (isDisabled) return;

    actions?.onChange?.(firstFile(event.dataTransfer?.files));
  }
</script>

<div class={cn('flex flex-col gap-1.5', ui?.className)}>
  <span class="text-ink-700 text-sm font-medium" aria-hidden="true">
    {data.label}
    {#if data.isRequired}
      <span class="text-destructive">*</span>
    {/if}
  </span>

  <label
    for={inputId}
    ondragover={(event) => {
      event.preventDefault();
      isDragging = !isDisabled;
    }}
    ondragleave={() => (isDragging = false)}
    ondrop={onDrop}
    class={cn(
      'control-lg rounded-control bg-muted flex w-full items-center gap-2 border-0 border-b text-sm transition-colors',
      'has-[:focus-visible]:ring-ring has-[:focus-visible]:ring-2',
      isDisabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer',
      isDragging && 'bg-primary-soft',
      hasError ? 'border-destructive' : 'border-input',
    )}
  >
    <Paperclip class="text-ink-500 size-4 shrink-0" aria-hidden="true" />
    <span class={cn('min-w-0 flex-1 truncate', !data.fileName && 'text-ink-500')}>
      {data.fileName ?? data.placeholder ?? 'Escolha ou arraste um arquivo'}
    </span>
    <span class="text-ink-700 shrink-0 text-xs font-medium">
      {data.fileName ? 'Trocar' : 'Escolher'}
    </span>
    <input
      id={inputId}
      name={data.name}
      type="file"
      accept={data.accept}
      autocomplete="off"
      disabled={isDisabled}
      required={data.isRequired}
      aria-label={data.label}
      aria-invalid={hasError}
      aria-describedby={hasError || data.hint ? messageId : undefined}
      onchange={(event) => actions?.onChange?.(firstFile(event.currentTarget.files))}
      onblur={() => actions?.onBlur?.()}
      class="sr-only"
    />
  </label>

  {#if hasError}
    <p id={messageId} role="alert" class="text-destructive text-xs font-medium">{error}</p>
  {:else if data.hint}
    <p id={messageId} class="text-muted-foreground text-xs">{data.hint}</p>
  {/if}
</div>
