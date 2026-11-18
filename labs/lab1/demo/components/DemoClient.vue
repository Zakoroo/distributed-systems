<script setup>
import { computed, onUnmounted, ref, watch } from 'vue'
import { demo, liveDemo, serverOrigin, validPort } from '../lib/demoState'

const props = defineProps({ broker: Boolean, deployment: Boolean })
const server = demo[props.deployment ? 'deployment' : 'local']
const clientId = ref('')
const selectedClient = computed(() => demo.clients.find(client => client.id === clientId.value))
watch(() => demo.clients.map(client => client.id), ids => {
  if (!ids.includes(clientId.value)) clientId.value = ids[0] ?? ''
}, { immediate: true })

const filePath = ref(props.deployment ? 'deployment-demo.txt' : 'demo.txt')
const topic = ref('weather')
const subscriptions = ref({})
const payloadType = ref('text')
const payload = ref(props.broker ? 'Rain at noon' : 'Hello from Slidev')
const upload = ref(null)
const requestBody = computed(() => payloadType.value === 'file' ? upload.value : payload.value)
const hasPayload = computed(() => requestBody.value !== null)
const pending = ref(new Set())
const results = ref([])

const origin = computed(() => serverOrigin(server))
const subscriptionKey = computed(() => JSON.stringify([origin.value, topic.value, clientId.value]))
const subscriberId = computed({
  get: () => subscriptions.value[subscriptionKey.value] ?? '',
  set: id => { subscriptions.value[subscriptionKey.value] = id },
})
const canSend = computed(() => server.host && validPort(server.port)
  && selectedClient.value?.address && validPort(selectedClient.value.port, true))
const busy = action => pending.value.has(clientId.value + ':' + action)
const clientBusy = computed(() => [...pending.value].some(key => key.startsWith(clientId.value + ':')))

function pickUpload(event) {
  upload.value = event.target.files[0] ?? null
  if (upload.value && !props.broker) filePath.value = upload.value.name
}

function clearResults() {
  for (const result of results.value) if (result.download) URL.revokeObjectURL(result.download)
  results.value = []
}

let disposed = false
onUnmounted(() => {
  disposed = true
  clearResults()
})

async function preview(data) {
  const bytes = await data.slice(0, 4096).arrayBuffer()
  try {
    const text = new TextDecoder('utf-8', { fatal: true }).decode(bytes, { stream: data.size > 4096 })
    if (!/[\x00-\x08\x0B\x0C\x0E-\x1F]/.test(text)) {
      return text + (data.size > 4096 ? '\n…' : '')
    }
  } catch {
    // Binary bytes are available through the download link.
  }
  return data.size + ' bytes of binary data'
}

async function send(action, method, path, body, filename = '') {
  const client = selectedClient.value
  if (!canSend.value || !client || busy(action)) return

  const key = client.id + ':' + action
  const subscription = subscriptionKey.value
  const label = client.name + ' · ' + method + ' ' + origin.value + path
  pending.value.add(key)
  try {
    const response = await fetch('/lab-demo/request', {
      method,
      body,
      headers: {
        'X-Lab-Demo': '1',
        'X-Lab-Origin': origin.value,
        'X-Lab-Path': path,
        'X-Lab-Source-Address': client.address,
        'X-Lab-Source-Port': String(client.port),
      },
      signal: AbortSignal.timeout(24000),
    })
    if (response.headers.has('X-Lab-Demo-Error')) throw new Error(await response.text())
    const source = response.headers.get('X-Lab-Source-Address')
    if (!source) throw new Error('Open the local Slidev development deck to send live requests.')

    let data = await response.blob()
    let text = ''
    if (action === 'Subscribe' && response.status === 201) {
      text = await data.text()
      subscriptions.value[subscription] = String(JSON.parse(text).id)
    } else if (action === 'Poll' && response.status === 200) {
      const event = JSON.parse(await data.text())
      data = new Blob([Uint8Array.from(atob(event.data), character => character.charCodeAt(0))])
      text = 'Event #' + event.sequence + '\n' + await preview(data)
      filename = 'event-' + event.sequence + '.bin'
    } else if (data.size) {
      text = await preview(data)
    }
    if (action === 'Unsubscribe' && response.ok) delete subscriptions.value[subscription]
    if (disposed) return

    const download = filename && response.ok && response.status !== 204 ? URL.createObjectURL(data) : ''
    results.value.unshift({
      label,
      text: 'HTTP ' + response.status + ' · source ' + source + ' port ' + response.headers.get('X-Lab-Source-Port')
        + (text ? '\n' + text : ''),
      download,
      filename,
    })
  } catch (error) {
    if (!disposed) results.value.unshift({ label, text: error.message })
  } finally {
    pending.value.delete(key)
  }
}

function sendFile(action, method) {
  const parts = filePath.value.split('/').filter(Boolean)
  if (!parts.length) return
  const path = '/' + parts.map(encodeURIComponent).join('/')
  return send(action, method, path, method === 'POST' ? requestBody.value : undefined,
    method === 'GET' ? parts.at(-1) : '')
}

function brokerAction(action, route, method) {
  const query = new URLSearchParams({ topic: topic.value })
  if (action === 'Poll' || action === 'Unsubscribe') query.set('id', subscriberId.value)
  return send(action, method, '/' + route + '?' + query, action === 'Publish' ? requestBody.value : undefined)
}
</script>

<template>
  <p v-if="!liveDemo">Start the Slidev development server to use the live client.</p>
  <section v-else class="demo-client" @keydown.stop @keyup.stop>
    <div class="demo-summary">
      <span>Server <code>{{ origin }}</code></span>
      <span v-if="selectedClient">{{ selectedClient.address }} · {{ selectedClient.port || 'automatic port' }}</span>
    </div>
    <div class="demo-fields">
      <label>Client
        <select v-model="clientId">
          <option v-for="client in demo.clients" :key="client.id" :value="client.id">{{ client.name || client.address }}</option>
        </select>
      </label>
      <label v-if="broker">Topic <input v-model.trim="topic" /></label>
      <label v-else>File path <input v-model.trim="filePath" /></label>
      <label v-if="broker">Subscriber ID <input v-model.trim="subscriberId" :disabled="clientBusy" placeholder="Filled on subscribe" /></label>
      <label>Payload type
        <select v-model="payloadType">
          <option value="text">Plain text</option>
          <option value="file">File upload</option>
        </select>
      </label>
      <label v-show="payloadType === 'text'">Text payload <input v-model="payload" /></label>
      <label v-show="payloadType === 'file'">File payload <input type="file" @change="pickUpload" /></label>
    </div>
    <div class="demo-actions">
      <template v-if="broker">
        <button :disabled="!canSend || clientBusy || !topic" @click="brokerAction('Subscribe', 'subscribe', 'POST')">Subscribe</button>
        <button :disabled="!canSend || !topic || !subscriberId || clientBusy" @click="brokerAction('Poll', 'events', 'GET')">Poll for event</button>
        <button :disabled="!canSend || !topic || clientBusy || !hasPayload" @click="brokerAction('Publish', 'publish', 'POST')">Publish</button>
        <button :disabled="!canSend || !topic || !subscriberId || clientBusy" @click="brokerAction('Unsubscribe', 'unsubscribe', 'POST')">Unsubscribe</button>
      </template>
      <template v-else>
        <button :disabled="!canSend || !filePath || clientBusy || !hasPayload" @click="sendFile('Append', 'POST')">POST / append</button>
        <button :disabled="!canSend || !filePath || clientBusy" @click="sendFile('Read', 'GET')">GET / read</button>
      </template>
      <button class="secondary" @click="clearResults">Clear output</button>
      <span v-if="pending.size" class="setup-help">{{ pending.size }} {{ pending.size === 1 ? 'request' : 'requests' }} in progress</span>
    </div>
    <div class="demo-output" aria-live="polite">
      <div v-for="(result, index) in results" :key="index">
        <strong>{{ result.label }}</strong>
        <pre>{{ result.text }}</pre>
        <a v-if="result.download" :href="result.download" :download="result.filename">Download {{ result.filename }}</a>
      </div>
      <p v-if="!results.length">Responses appear here.</p>
    </div>
  </section>
</template>
