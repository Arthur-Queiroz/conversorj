# CLAUDE.md

Contexto e convenções do projeto **ConversorJ**. Leia antes de qualquer alteração.

## O que é

Conversor web self-hosted de vídeos do **YouTube e X (Twitter)** para **MP3 ou MP4** (formato à escolha do usuário). Sem cadastro, sem anúncios, sem persistência de dados pessoais. Uso interno/pessoal, deploy em VPS própria.

A especificação funcional completa está em `conversorj-spec.md`. Em caso de conflito, este arquivo manda nas convenções de código; a spec manda no comportamento do produto.

## Stack

- **Back-end:** Go + GORM, API REST. SQLite para estado de jobs.
- **Conversão:** `yt-dlp` + `ffmpeg` invocados via exec.
- **Front-end:** Vue 3 (estático, sem SSR) + Tailwind CSS.
- **Deploy:** Docker Compose, VPS KVM2 Hostinger, HTTPS via reverse proxy.
- **Redis:** NÃO usar no MVP. Só introduzir se filas virarem necessidade real.

## Estrutura de pastas

```
/backend        API Go
  /internal
    /handler    handlers HTTP
    /converter  wrapper yt-dlp/ffmpeg
    /validator  validação de URL por plataforma
    /model      structs GORM
  main.go
/frontend       Vue + Tailwind
docker-compose.yml
```

## Convenções de código

### Go
- `gofmt`/`goimports` sempre. Sem código não formatado.
- Erros: retornar `error` explícito, nunca `panic` em fluxo normal. Embrulhar com `fmt.Errorf("...: %w", err)`.
- Handlers retornam JSON; erros de validação são `400` com `{ "error": "código", "message": "texto pt-BR" }`.
- Nada de variáveis globais mutáveis; injetar dependências (DB, config) via struct.
- Nomes de pacotes curtos e minúsculos.

### Front-end
- Vue 3 **Composition API** com `<script setup>`.
- Tailwind para tudo; evitar CSS custom salvo necessidade.
- Visual minimalista, claro e elegante. Foco na ação principal.
- Responsivo (mobile-first), **não** é PWA.
- Mensagens de UI em **português do Brasil**.
- **Cliente HTTP: `axios`** (não usar `fetch`). Criar uma instância única (`src/api/client.js` ou similar) com `baseURL: '/api'` e timeout configurado; importar essa instância nos componentes em vez de chamar axios direto.

## Comportamento crítico

- **Formato é escolha do usuário:** a UI deve ter seletor MP3/MP4 e a API recebe `format` no payload. Nunca assumir um formato fixo.
- **Validar URL por plataforma** antes de chamar o conversor (YouTube e X apenas — ver tabela na spec, seção 9).
- **Limite de duração: 20 minutos.** Verificar a duração via metadados do `yt-dlp` ANTES de baixar. Vídeos acima disso retornam `400` com `{ "error": "duration_exceeded", ... }`.
- **Rate limit por IP:** 10 conversões/hora → `429` com `{ "error": "rate_limit_exceeded", ... }`.
- **Concorrência global:** máximo de 5 conversões simultâneas (semáforo). Ao exceder, retornar imediatamente `429` com `{ "error": "server_busy", "message": "Muitos vídeos estão sendo processados no momento. Aguarde um instante e tente novamente." }`. Não enfileirar.
- **yt-dlp sempre atualizado.** Imagem Docker deve buscar a versão mais recente no build; considerar job de atualização. yt-dlp desatualizado quebra a extração.
- **Sem dados de usuário.** Não logar URLs atreladas a identidade. Arquivos temporários removidos após download ou TTL curto.
- **Conversão síncrona.** `POST /api/convert` processa e retorna um `downloadUrl` na própria resposta. NÃO implementar fila de jobs nem endpoints de status no MVP.
- **Timeout** em toda conversão para não travar recursos.

## API (resumo)

- `POST /api/convert` → `{ platform, url, format }` (`platform`: `youtube` | `x`) → retorna `{ downloadUrl, filename, expiresAt }`
- `GET /api/download/:id` → arquivo (removido após download / TTL)

Detalhes e exemplos em `conversorj-spec.md`, seção 6.

## O que NÃO fazer

- Não adicionar Redis, autenticação ou banco de usuários sem pedido explícito.
- Não introduzir build pesado no front-end nem dependências supérfluas.
- Não persistir dados pessoais.
- Não tornar a aplicação PWA.
- Não fixar o formato de saída — sempre respeitar a escolha do usuário.
- Não suportar Instagram (fora de escopo).
- Não pular a checagem de duração de 20 min antes do download.
- Não usar APIs oficiais do YouTube/X para download — não funcionam para isso; usar yt-dlp.

## Idioma

Código, nomes de variáveis e commits em inglês. Texto voltado ao usuário (UI, mensagens de erro) em português do Brasil.
