# ConversorJ

Conversor web self-hosted de vídeos do YouTube e X (Twitter) para MP3 ou MP4.

## Requisitos

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (inclui Docker Compose)
- [Git](https://git-scm.com/)

## Como rodar localmente

**1. Clone o repositório**

```bash
git clone https://github.com/Arthur-Queiroz/conversorj.git
cd conversorj
```

**2. Crie o arquivo `.env`**

```bash
echo "DOMAIN=http://localhost" > .env
```

**3. Suba os containers**

```bash
docker compose up -d
```

O primeiro build demora alguns minutos (baixa dependências e compila o frontend). Os próximos iniciam em segundos.

**4. Acesse no navegador**

```
http://localhost
```

## Parar a aplicação

```bash
docker compose down
```

## Atualizar após mudanças no código

```bash
git pull
docker compose up -d --build
```

---

> **Nota:** A aplicação foi projetada para rodar localmente com IP residencial.
> Rodar em VPS de data center (Hostinger, DigitalOcean, OVH, etc.) causa erro 403
> ao tentar baixar vídeos, pois o YouTube bloqueia downloads nesses IPs.
