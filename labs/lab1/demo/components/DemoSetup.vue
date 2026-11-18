<script setup>
import { demo, liveDemo } from '../lib/demoState'

const props = defineProps({ deployment: Boolean })
const server = demo[props.deployment ? 'deployment' : 'local']

function addClient() {
  demo.clients.push({
    id: crypto.randomUUID(),
    name: 'Client ' + (demo.clients.length + 1),
    address: '127.0.0.1',
    port: 0,
  })
}
</script>

<template>
  <p v-if="!liveDemo">Open the deck in Slidev development mode to configure the live demo.</p>
  <section v-else class="demo-setup" @keydown.stop @keyup.stop>
    <div class="server-fields">
      <label>Protocol <select v-model="server.scheme"><option value="http">HTTP</option><option value="https">HTTPS</option></select></label>
      <label>Server address <input v-model.trim="server.host" placeholder="127.0.0.1" /></label>
      <label>Server port <input v-model.number="server.port" type="number" min="1" max="65535" /></label>
    </div>
    <template v-if="!deployment">
      <h2>Clients</h2>
      <div class="client-list">
        <div v-for="(client, index) in demo.clients" :key="client.id" class="client-fields">
          <label>Alias <input v-model.trim="client.name" /></label>
          <label>Source IP <input v-model.trim="client.address" /></label>
          <label>Source port <input v-model.number="client.port" type="number" min="0" max="65535" /></label>
          <button type="button" class="secondary" :disabled="demo.clients.length === 1" @click="demo.clients.splice(index, 1)">Remove</button>
        </div>
      </div>
      <button type="button" @click="addClient">Add client</button>
      <p class="setup-help">Use an IP assigned to this machine. Port 0 lets the OS pick a free port. Use distinct fixed ports for clients that send requests at the same time.</p>
    </template>
    <p v-else class="setup-help">For an SSH tunnel, enter its local address and port here. For a directly reachable deployment, enter the remote HTTP address and port. The live deployment check uses this target with the same client list.</p>
  </section>
</template>
