<script lang="ts">
  import { cn } from '$lib/utils/cn';

  export type AcerolaBrandMarkProps = {
    ui?: {
      size?: 'sm' | 'md' | 'lg';
      className?: string;
    };
  };

  /* A marca é SÓ a logo, sem o nome escrito ao lado — a mesma dos documentos exportados. O
     arquivo é o ícone da aba (`static/favicon.svg`): uma fonte só para a marca na tela. É o
     selo padrão dos apps — quadrado na primária, monograma e a faixa no destaque. */
  const LOGO_SIZES = { sm: 'size-10', md: 'size-12', lg: 'size-20' } as const;

  let { ui }: AcerolaBrandMarkProps = $props();
</script>

<!-- O nome vai no texto alternativo: quem usa leitor de tela continua ouvindo qual é o sistema. -->
<img
  src="/favicon.svg"
  alt="artifact-docs"
  class={cn('shrink-0 object-contain', LOGO_SIZES[ui?.size ?? 'md'], ui?.className)}
/>
