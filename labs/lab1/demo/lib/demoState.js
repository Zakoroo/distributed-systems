import { reactive, watch } from 'vue'

const key = 'lab1-slidev-demo-v2'
export const liveDemo = import.meta.env.DEV

const defaults = {
  local: { scheme: 'http', host: '127.0.0.1', port: 8080 },
  deployment: { scheme: 'http', host: '127.0.0.1', port: 9080 },
  clients: [
    { id: 'alice', name: 'Alice', address: '127.0.0.1', port: 0 },
    { id: 'bob', name: 'Bob', address: '127.0.0.1', port: 0 },
  ],
}

let saved = {}
try {
  saved = JSON.parse(globalThis.localStorage?.getItem(key) ?? '{}') ?? {}
} catch {
  // A malformed or unavailable storage entry leaves the defaults usable.
}
export const demo = reactive({ ...defaults, ...saved })

watch(demo, () => {
  try {
    globalThis.localStorage?.setItem(key, JSON.stringify(demo))
  } catch {
    // The current presentation still has the configuration in memory.
  }
})

globalThis.addEventListener?.('storage', event => {
  if (event.key !== key || !event.newValue) return
  try {
    const saved = JSON.parse(event.newValue)
    // Keep the server objects that already-mounted slides are using.
    Object.assign(demo.local, saved.local)
    Object.assign(demo.deployment, saved.deployment)
    demo.clients = saved.clients
  } catch {
    // Leave this window's configuration available if storage was edited.
  }
})

export function formatHost(host) {
  const name = host.trim()
  return name.includes(':') && !name.startsWith('[') ? '[' + name + ']' : name
}

export function serverOrigin(server) {
  return server.scheme + '://' + formatHost(server.host) + ':' + server.port
}

export function validPort(port, allowAuto = false) {
  return Number.isInteger(port) && port >= (allowAuto ? 0 : 1) && port <= 65535
}
