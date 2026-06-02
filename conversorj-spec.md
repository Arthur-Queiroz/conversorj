# ConversorJ — Especificação Técnica

> Conversor de vídeos do YouTube e X (Twitter) para MP3/MP4. Self-hosted, leve, sem anúncios e sem dependências de serviços de terceiros questionáveis.

---

## 1. Visão Geral

### Problema
Converter vídeos do YouTube para arquivos de mídia hoje exige passar por sites e apps de procedência duvidosa, cheios de anúncios, popups e riscos de segurança.

### Nota técnica importante
Este tipo de aplicação **não usa APIs oficiais** do YouTube ou X — elas não fornecem download de mídia e proíbem isso nos termos de uso. A conversão é feita por **extração** via `yt-dlp`, que faz engenharia reversa dos players e localiza o stream real, seguida de transcodificação com `ffmpeg`. Isso implica fragilidade (players mudam) e necessidade de manutenção contínua do `yt-dlp`. Ver seção 13.

### Solução
Uma aplicação web interna, simples e eficiente, que faz a conversão de forma direta: o usuário escolhe a plataforma, cola o link, clica em converter e baixa o arquivo. Sem cadastro, sem ruído, sem anúncios.

### Princípios de design
- **Leve**: front-end estático, sem build pesado nem estado complexo.
- **Minimalista**: visual claro, elegante, foco total na ação principal.
- **Sem persistência de usuário**: nenhum dado pessoal é armazenado.
- **Self-hosted**: roda inteiramente em infraestrutura própria.

---

## 2. Stack Técnica

| Camada | Tecnologia | Observações |
|---|---|---|
| Front-end | Vue 3 (estático) + Tailwind CSS + axios | UI simples, responsiva, sem SSR; axios como cliente HTTP |
| Back-end | Go + GORM | API REST |
| Banco de dados | SQLite | Apenas para metadados de jobs/cache, se necessário |
| Conversão de mídia | yt-dlp + ffmpeg | Binários invocados pelo back-end |
| Deploy | Docker / Docker Compose | VPS KVM2 da Hostinger |
| Cache/filas (opcional) | Redis | Ver seção 7 |

> **Nota sobre Redis:** Para um MVP de uso pessoal/interno, Redis **não é necessário**. SQLite resolve o controle de estado dos jobs. Redis só passa a fazer sentido se houver muitas conversões concorrentes e for preciso uma fila de trabalho dedicada (ver seção 7).

---

## 3. Requisitos Funcionais

| ID | Requisito | Prioridade |
|---|---|---|
| RF-01 | O usuário deve poder selecionar a plataforma de origem: YouTube ou X.com | Alta |
| RF-02 | O usuário deve poder colar um link e clicar em "Converter" | Alta |
| RF-03 | Após conversão, o navegador deve retornar o arquivo (MP3 ou MP4) para download | Alta |
| RF-04 | URLs inválidas devem alertar o usuário com mensagem clara | Alta |
| RF-05 | O usuário escolhe o formato de saída (MP3 ou MP4) via seletor na UI | Alta |
| RF-06 | Estado de carregamento/progresso deve ser exibido durante a conversão | Média |
| RF-07 | Vídeos com duração acima de **20 minutos** devem ser rejeitados, alertando o usuário antes de iniciar o download | Alta |

---

## 4. Requisitos Não-Funcionais

- **Sem PWA**, mas **web responsivo** — utilizável em desktop e celular.
- **Sem autenticação nem dados de usuário**. Se cookies forem usados (ex.: preferência de tema), exibir banner de consentimento.
- **Leveza**: bundle de front-end mínimo, sem dependências desnecessárias.
- **Privacidade**: nenhum log de URLs convertidas atrelado a usuário.

---

## 5. Arquitetura

```
┌─────────────┐      HTTPS       ┌──────────────────┐
│  Vue + TW   │ ───────────────► │   API Go (GORM)  │
│  (estático) │ ◄─────────────── │                  │
└─────────────┘   JSON / file    └────────┬─────────┘
                                           │ exec
                                  ┌────────▼─────────┐
                                  │  yt-dlp + ffmpeg │
                                  └────────┬─────────┘
                                           │
                                  ┌────────▼─────────┐
                                  │  SQLite (jobs)   │
                                  └──────────────────┘
```

### Fluxo de conversão
1. Front-end envia `POST /api/convert` com `{ platform, url, format }`.
2. Back-end valida a URL contra o padrão da plataforma escolhida.
3. Back-end consulta os **metadados** via `yt-dlp` (sem baixar) e verifica a duração; se > 20 min, retorna erro.
4. Back-end invoca `yt-dlp` (+ `ffmpeg` para extração/transcodificação).
5. Arquivo gerado é devolvido como download (stream) ou via URL temporária.
6. Arquivo temporário é removido após o download / TTL.

### Camadas de rede (importante)
- **Front-end → API própria:** o Vue chama apenas a própria API (`/api/...`) usando **axios**. Nenhuma chamada do front-end vai direto às plataformas.
- **Back-end → plataformas:** o back-end Go **não faz requisições HTTP de saída** para YouTube/X. Ele invoca `yt-dlp` como processo local (exec), que cuida de toda a comunicação com as plataformas. Não reimplementar extração de mídia em Go.

---

## 6. API REST

> **Modelo síncrono.** A conversão acontece dentro da própria requisição `POST /api/convert`, que devolve um link de download temporário ao terminar. Não há endpoints de job no MVP. Ver seção 6.1 para a evolução assíncrona futura.

### `POST /api/convert`
Executa a conversão e retorna o link do arquivo. A requisição permanece aberta durante o processamento (configurar timeout generoso no reverse proxy e no axios).

**Request**
```json
{
  "platform": "youtube",   // "youtube" | "x"
  "url": "https://...",
  "format": "mp3"          // "mp3" | "mp4"
}
```

**Response (sucesso)**
```json
{
  "downloadUrl": "/api/download/abc123",
  "filename": "video.mp3",
  "expiresAt": "2025-06-02T12:00:00Z"
}
```

**Response (erro de validação)**
```json
{
  "error": "invalid_url",
  "message": "O link informado não é válido para a plataforma YouTube."
}
```

**Response (vídeo muito longo)**
```json
{
  "error": "duration_exceeded",
  "message": "O vídeo tem mais de 20 minutos e não pode ser convertido."
}
```

**Response (limite por IP atingido — HTTP 429)**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Você atingiu o limite de conversões por hora. Tente novamente mais tarde."
}
```

**Response (capacidade momentânea esgotada — HTTP 429)**
```json
{
  "error": "server_busy",
  "message": "Muitos vídeos estão sendo processados no momento. Aguarde um instante e tente novamente."
}
```

### `GET /api/download/:id`
Faz o download do arquivo convertido. O arquivo é removido após o download ou ao expirar (TTL).

> **Por que devolver um link em vez do arquivo no corpo?** Mantém a resposta da conversão leve e desacopla o download, facilitando uma futura migração para assíncrono sem reescrever o front-end.

### 6.1. Evolução futura: assíncrono
Se no futuro surgir necessidade de barra de progresso real ou houver problemas de timeout sob carga, o modelo pode evoluir para assíncrono: `POST /api/convert` passaria a enfileirar e retornar um `jobId` + `status`, com um `GET /api/jobs/:id` para polling. Isso exigiria a tabela `Job` com estados, um worker de fila e (possivelmente) Redis. **Fora do escopo do MVP.**


---

## 7. Decisão: Redis é necessário?

| Cenário | Redis? |
|---|---|
| Uso pessoal/interno, poucas conversões | ❌ Não — SQLite basta |
| Conversões longas com fila e workers | ✅ Sim — fila de jobs |
| Cache de URLs já convertidas | 🟡 Opcional — pode usar SQLite |

**Recomendação:** começar **sem Redis**. Adicionar depois apenas se a fila virar gargalo.

---

## 8. Modelo de Dados (SQLite)

No modelo síncrono não há fila de jobs. O SQLite serve para rastrear os arquivos gerados e sua expiração (para o endpoint de download e a limpeza por TTL).

```go
type Conversion struct {
    ID        string    `gorm:"primaryKey"` // usado na downloadUrl
    Platform  string
    Format    string    // mp3 | mp4
    Filename  string
    FilePath  string
    CreatedAt time.Time
    ExpiresAt time.Time // limpeza automática
}
```

> Se preferir simplicidade máxima, o estado pode até ser mantido só em memória/disco sem SQLite. O SQLite é recomendado para sobreviver a reinícios do container e facilitar a limpeza por TTL.

---

## 9. Validação de URL por plataforma

| Plataforma | Padrão esperado |
|---|---|
| YouTube | `youtube.com/watch?v=`, `youtu.be/` |
| X (Twitter) | `x.com/.../status/`, `twitter.com/.../status/` |

---

## 9.1. Rate Limiting e Concorrência

Dois limites independentes protegem a API contra abuso e protegem a VPS de sobrecarga. Os valores são propositalmente baixos, condizentes com um projeto de pouquíssimos usuários.

| Limite | Valor | Escopo | Resposta ao exceder |
|---|---|---|---|
| Conversões por IP | **10 por hora** | Por endereço IP | `429` com `rate_limit_exceeded` |
| Conversões simultâneas | **5 no total** | Global (toda a API) | `429` com `server_busy` |

- **Por IP (10/hora):** middleware contando conversões por IP em janela de 1 hora. Folgado para uso humano normal, mas corta scripts abusivos.
- **Concorrência global (máx. 5):** controlada por um semáforo. Quando 5 conversões já estão em andamento, novas requisições recebem imediatamente o erro `server_busy` (não enfileirar nem bloquear a requisição). A mensagem ao usuário deve ser: *"Muitos vídeos estão sendo processados no momento. Aguarde um instante e tente novamente."*

> A concorrência global é a proteção mais importante: cada conversão consome CPU intensa (ffmpeg) e a KVM2 não suporta muitas em paralelo.

---

## 10. Deploy

- **Docker Compose** com um serviço para a API Go (com `yt-dlp` e `ffmpeg` instalados na imagem) e um para servir o front-end estático (ou servido pela própria API/Nginx).
- VPS **KVM2 da Hostinger**.
- HTTPS via reverse proxy (Nginx/Caddy/Traefik).
- Volume persistente para o SQLite e diretório temporário de mídia com limpeza periódica.

---

## 11. Roadmap de Implementação (para Claude Code)

> Ordem sugerida de tarefas, cada uma autocontida.

1. **Setup do back-end Go** — módulo, estrutura de pastas, GORM + SQLite, healthcheck.
2. **Wrapper de conversão** — função que (a) consulta metadados via `yt-dlp` e checa duração (≤ 20 min), e (b) invoca `yt-dlp`/`ffmpeg` retornando o caminho do arquivo.
3. **Validação de URL** — por plataforma (regex), YouTube e X.
4. **Endpoint `POST /api/convert`** — síncrono, retorna `downloadUrl`. Inclui erros `invalid_url` e `duration_exceeded`.
5. **Endpoint `GET /api/download/:id`** — serve o arquivo e respeita o TTL.
6. **Rate limit + concorrência** — middleware de 10 conversões/hora por IP (`rate_limit_exceeded`) e semáforo global de 5 conversões simultâneas (`server_busy`). Ambos retornam `429`.
7. **Front-end Vue + Tailwind + axios** — seletor de plataforma, input de URL, seletor de formato, botão converter, estados de loading/erro.
8. **Integração e download** — `axios.post` → recebe `downloadUrl` → dispara o download.
9. **Dockerfile + docker-compose** — imagem com `yt-dlp`/`ffmpeg` (sempre atualizados no build), reverse proxy com timeout generoso.
10. **Limpeza de arquivos temporários** — job periódico / TTL.
11. **(Opcional) Banner de cookies, tema claro/escuro.**

---

## 12. Decisões Tomadas

Todos os pontos em aberto foram resolvidos:

- **Formato:** MP3 ou MP4, à escolha do usuário.
- **Plataformas:** YouTube e X (Instagram fora de escopo).
- **Duração máxima:** 20 minutos (verificada antes do download).
- **Modelo:** síncrono (assíncrono fica como evolução futura — seção 6.1).
- **Rate limit por IP:** 10 conversões/hora.
- **Concorrência global:** máximo de 5 conversões simultâneas.
- **Cliente HTTP do front:** axios.

---

## 13. Limitações e Restrições Técnicas

Restrições inerentes a este tipo de aplicação, já incorporadas aos requisitos:

### Extração, não API oficial
- A conversão depende do **`yt-dlp`**, que extrai mídia por engenharia reversa dos players. Não há API oficial que entregue o arquivo.
- **Manutenção contínua obrigatória:** quando as plataformas mudam, a extração quebra. O `yt-dlp` deve ser **atualizado automaticamente** (no build/deploy ou via job periódico). Um yt-dlp desatualizado para de funcionar em semanas.

### Bloqueio de IP (YouTube)
- O YouTube detecta tráfego de datacenters/VPS e pode exigir verificação anti-bot. Para uso interno de baixo volume costuma passar.
- Mitigações disponíveis se necessário: limitar volume, suporte opcional a cookies de sessão, proxies. Não implementar de início.

### Limites impostos pela aplicação
- **Duração máxima: 20 minutos** (RF-07), verificada via metadados antes do download.
- **Rate limit por IP** para evitar abuso e reduzir risco de bloqueio.
- **Timeout** de conversão para liberar recursos travados.
- **Limpeza agressiva** de arquivos temporários (TTL curto).

### Recursos da VPS
- `ffmpeg` é intensivo em CPU. Numa KVM2, conversões concorrentes de MP4 podem saturar o servidor.
- Recomendado limitar o número de conversões simultâneas. É exatamente o cenário onde uma fila com workers limitados (e, aí sim, Redis) passaria a fazer sentido (ver seção 7).

### Legal / Termos de Uso
- Baixar do YouTube viola os termos de uso da plataforma, mesmo para uso pessoal. Por isso apps comerciais do tipo vivem em zona cinzenta.
- Como uso interno e self-hosted, o risco prático é baixo, mas **não recomendado** expor publicamente ou associar formalmente ao MEI/negócio.
