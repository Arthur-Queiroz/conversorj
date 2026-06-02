import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  timeout: 300_000, // 5 min — conversões podem demorar
})

export default client
