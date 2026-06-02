<script setup>
import { ref, computed } from 'vue'
import client from '../api/client'

const state = ref('initial')
const platform = ref('youtube')
const format = ref('mp3')
const url = ref('')

const result = ref(null) // { downloadUrl, filename, expiresAt }

const busy = computed(() => state.value === 'loading')
const urlError = computed(() => state.value === 'url')

const platforms = {
  youtube: { label: 'YouTube', placeholder: 'https://youtube.com/watch?v=…' },
  x: { label: 'x.com (Twitter)', placeholder: 'https://x.com/usuario/status/…' },
}

const currentPlatform = computed(() => platforms[platform.value])

function setPlatform(p) {
  if (!busy.value) platform.value = p
}

async function convert() {
  if (!url.value.trim()) {
    state.value = 'url'
    return
  }
  state.value = 'loading'
  result.value = null
  try {
    const { data } = await client.post('/convert', {
      platform: platform.value,
      url: url.value.trim(),
      format: format.value,
    })
    result.value = data
    state.value = 'success'
  } catch (err) {
    const code = err.response?.data?.error
    if (code === 'duration_exceeded') state.value = 'duration'
    else if (code === 'rate_limit_exceeded') state.value = 'busy'
    else if (code === 'server_busy') state.value = 'busy'
    else if (code === 'invalid_url') state.value = 'url'
    else state.value = 'error'
  }
}

function restart() {
  state.value = 'initial'
  url.value = ''
  result.value = null
}
</script>

<template>
  <div class="wstage">
    <div class="wwrap">

      <!-- brand -->
      <div class="wbrand">
        <div class="mark">
          <svg viewBox="0 0 24 24" fill="none">
            <path d="M12 3.5v11m0 0 4-4M12 14.5l-4-4" stroke="#fff" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M6 17.5h12" stroke="#fff" stroke-width="2.2" stroke-linecap="round"/>
          </svg>
        </div>
        <div>
          <div class="brand-name">Conversor<b>J</b></div>
          <div class="brand-sub">YouTube e X → MP3 ou MP4, sem ruído.</div>
        </div>
      </div>

      <!-- wide card -->
      <div class="wcard">
        <div class="wgrid">

          <!-- left: form -->
          <div class="wform">

            <!-- step 01: platform -->
            <div class="block">
              <div class="field-label">
                <span class="step">01</span> Plataforma
              </div>
              <div class="seg">
                <button
                  :aria-pressed="platform === 'youtube'"
                  @click="setPlatform('youtube')"
                  :disabled="busy"
                >
                  <span class="pg" :style="{ color: platform === 'youtube' ? '#E33' : 'var(--ink-faint)' }">
                    <!-- YouTube icon -->
                    <svg viewBox="0 0 24 24" fill="none">
                      <rect x="2.5" y="5" width="19" height="14" rx="4" fill="currentColor"/>
                      <path d="M10 9.2v5.6l4.8-2.8L10 9.2Z" fill="#fff"/>
                    </svg>
                  </span>
                  YouTube
                </button>
                <button
                  :aria-pressed="platform === 'x'"
                  @click="setPlatform('x')"
                  :disabled="busy"
                >
                  <span class="pg" :style="{ color: platform === 'x' ? '#111' : 'var(--ink-faint)' }">
                    <!-- X/Twitter icon -->
                    <svg viewBox="0 0 24 24" fill="currentColor">
                      <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24h-6.66l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231 5.45-6.231Zm-1.161 17.52h1.833L7.084 4.126H5.117l11.966 15.644Z"/>
                    </svg>
                  </span>
                  x.com (Twitter)
                </button>
              </div>
            </div>

            <!-- step 02: url -->
            <div class="block" :class="{ dim: busy }">
              <div class="field-label">
                <span class="step">02</span> Link do vídeo
              </div>
              <div class="input-wrap" :class="{ error: urlError }">
                <span class="lead">
                  <svg viewBox="0 0 24 24" fill="none" width="18" height="18">
                    <path d="M9 15l6-6M10.5 6.5l1-1a4 4 0 0 1 6 6l-1 1M13.5 17.5l-1 1a4 4 0 0 1-6-6l1-1" stroke="currentColor" stroke-width="1.9" stroke-linecap="round"/>
                  </svg>
                </span>
                <input
                  type="text"
                  v-model="url"
                  :placeholder="currentPlatform.placeholder"
                  :disabled="busy"
                  @keyup.enter="convert"
                />
              </div>
              <div v-if="urlError" class="inline-err">
                <svg viewBox="0 0 24 24" fill="none">
                  <path d="M12 9v4.5M12 17h.01" stroke="currentColor" stroke-width="2.1" stroke-linecap="round"/>
                  <path d="M10.3 4.3 2.5 18a2 2 0 0 0 1.7 3h15.6a2 2 0 0 0 1.7-3L13.7 4.3a2 2 0 0 0-3.4 0Z" stroke="currentColor" stroke-width="1.9" stroke-linejoin="round"/>
                </svg>
                O link informado não é válido para a plataforma {{ currentPlatform.label }}.
              </div>
            </div>

            <!-- step 03: format -->
            <div class="block" :class="{ dim: busy }">
              <div class="field-label">
                <span class="step">03</span> Formato de saída
              </div>
              <div class="fmt">
                <label>
                  <input type="radio" name="wfmt" value="mp3" v-model="format" />
                  <span class="ficon">
                    <svg viewBox="0 0 24 24" fill="none">
                      <path d="M9 17V6l10-2v11" stroke="currentColor" stroke-width="1.9" stroke-linejoin="round"/>
                      <circle cx="6.5" cy="17.5" r="2.7" stroke="currentColor" stroke-width="1.9"/>
                      <circle cx="16.5" cy="15.5" r="2.7" stroke="currentColor" stroke-width="1.9"/>
                    </svg>
                  </span>
                  <span class="ftext">
                    <b>MP3</b>
                    <span>Apenas áudio</span>
                  </span>
                  <span class="check">
                    <svg viewBox="0 0 24 24" fill="none" width="17" height="17">
                      <path d="M5 12.5l4.5 4.5L19 7" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                  </span>
                </label>
                <label>
                  <input type="radio" name="wfmt" value="mp4" v-model="format" />
                  <span class="ficon">
                    <svg viewBox="0 0 24 24" fill="none">
                      <rect x="3" y="6" width="13" height="12" rx="2.4" stroke="currentColor" stroke-width="1.9"/>
                      <path d="M16 10.5 21 8v8l-5-2.5" stroke="currentColor" stroke-width="1.9" stroke-linejoin="round"/>
                    </svg>
                  </span>
                  <span class="ftext">
                    <b>MP4</b>
                    <span>Vídeo completo</span>
                  </span>
                  <span class="check">
                    <svg viewBox="0 0 24 24" fill="none" width="17" height="17">
                      <path d="M5 12.5l4.5 4.5L19 7" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                  </span>
                </label>
              </div>
            </div>

          </div>

          <!-- right: action panel -->
          <div class="wpanel">

            <!-- loading state -->
            <div v-if="state === 'loading'" class="loading">
              <div class="top">
                <div class="spinner"></div>
                <div class="lt">
                  <b>Convertendo…</b>
                  <span>extraindo {{ format === 'mp3' ? 'áudio' : 'vídeo' }} · ffmpeg</span>
                </div>
              </div>
              <div class="bar"><i></i></div>
            </div>

            <!-- success state -->
            <div v-else-if="state === 'success'" class="result">
              <div class="rhead">
                <span class="rcheck">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M5 12.5l4.5 4.5L19 7" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </span>
                <div>
                  <b>Conversão concluída</b>
                  <span>Seu arquivo está pronto.</span>
                </div>
              </div>
              <div class="file">
                <span class="fb">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M7 3.5h6.5L19 9v10.5a1.5 1.5 0 0 1-1.5 1.5h-10A1.5 1.5 0 0 1 6 19.5V5A1.5 1.5 0 0 1 7 3.5Z" stroke="currentColor" stroke-width="1.7"/>
                    <path d="M13 3.5V9h5" stroke="currentColor" stroke-width="1.7" stroke-linejoin="round"/>
                  </svg>
                </span>
                <div class="fn">
                  <b>{{ result?.filename ?? '' }}</b>
                  <span>{{ format === 'mp3' ? 'Áudio · 320 kbps' : 'Vídeo' }}</span>
                </div>
                <span class="badge">.{{ format }}</span>
              </div>
              <a class="dl" :href="result?.downloadUrl" download>
                <svg viewBox="0 0 24 24" fill="none">
                  <path d="M12 4v11m0 0 4.5-4.5M12 15l-4.5-4.5" stroke="currentColor" stroke-width="2.1" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M5 19h14" stroke="currentColor" stroke-width="2.1" stroke-linecap="round"/>
                </svg>
                Baixar arquivo
              </a>
              <div class="expire">
                <svg viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="8.4" stroke="currentColor" stroke-width="1.9"/>
                  <path d="M12 7.5V12l3 2" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                O link expira em 10 minutos.
              </div>
              <button class="restart" @click="restart">Nova conversão</button>
            </div>

            <!-- default: ready state (with possible alerts) -->
            <template v-else>
              <div class="wp-kicker">Pronto para converter</div>
              <div class="wp-fmt">Saída em <em>{{ format === 'mp3' ? 'MP3' : 'MP4' }}</em></div>
              <div class="wp-sub">{{ format === 'mp3' ? 'Áudio de alta qualidade, 320 kbps.' : 'Vídeo completo em até 1080p.' }}</div>

              <div v-if="state === 'duration'" class="alert warn">
                <span class="ai">
                  <svg viewBox="0 0 24 24" fill="none">
                    <circle cx="12" cy="12" r="8.4" stroke="currentColor" stroke-width="1.9"/>
                    <path d="M12 7.5V12l3 2" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </span>
                <div>
                  <b>Vídeo muito longo</b>
                  Mais de 20 minutos. Tente um trecho mais curto.
                </div>
              </div>

              <div v-if="state === 'busy'" class="alert busy">
                <span class="ai">
                  <svg viewBox="0 0 24 24" fill="none">
                    <path d="M12 3a9 9 0 1 0 9 9" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
                    <path d="M12 7v5l3 2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
                </span>
                <div>
                  <b>Servidor ocupado</b>
                  Aguarde um instante e tente novamente.
                </div>
              </div>

              <button class="cta" @click="convert" :disabled="busy">
                <svg viewBox="0 0 24 24" fill="none">
                  <path d="M12 3v12m0 0 4.5-4.5M12 15l-4.5-4.5" stroke="currentColor" stroke-width="2.1" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M5 17v1.5A1.5 1.5 0 0 0 6.5 20h11a1.5 1.5 0 0 0 1.5-1.5V17" stroke="currentColor" stroke-width="2.1" stroke-linecap="round"/>
                </svg>
                Converter {{ format === 'mp3' ? 'em MP3' : 'em MP4' }}
              </button>

              <div class="whints">
                <span class="hint">
                  <svg viewBox="0 0 24 24" fill="none"><circle cx="12" cy="12" r="8.4" stroke="currentColor" stroke-width="1.9"/><path d="M12 7.5V12l3 2" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"/></svg>
                  Máx. 20 min por vídeo
                </span>
              </div>
            </template>

          </div>
        </div>
      </div>

    </div>
  </div>
</template>
