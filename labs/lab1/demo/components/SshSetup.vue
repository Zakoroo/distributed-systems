<script setup>
import { computed, reactive, ref } from 'vue'
import { demo, formatHost, liveDemo, validPort } from '../lib/demoState'

const ssh = reactive({
  user: 'ubuntu',
  host: '',
  port: 22,
  keyPath: '',
  localPort: 9080,
  remoteHost: '127.0.0.1',
  remotePort: 8080,
})
const copied = ref(false)
const configured = computed(() => ssh.user && ssh.host && ssh.remoteHost
  && [ssh.port, ssh.localPort, ssh.remotePort].every(port => validPort(port)))
const quote = value => "'" + String(value).replaceAll("'", "'\"'\"'") + "'"
const command = computed(() => {
  const forward = '127.0.0.1:' + ssh.localPort + ':' + formatHost(ssh.remoteHost) + ':' + ssh.remotePort
  const key = ssh.keyPath ? ' -i ' + quote(ssh.keyPath) : ''
  return 'ssh -N -o ExitOnForwardFailure=yes -p ' + ssh.port + key
    + ' -L ' + quote(forward) + ' -- ' + quote(ssh.user + '@' + ssh.host)
})
function useTunnel() {
  Object.assign(demo.deployment, { scheme: 'http', host: '127.0.0.1', port: ssh.localPort })
}
async function copyCommand() {
  try {
    await navigator.clipboard.writeText(command.value)
    copied.value = true
  } catch {
    copied.value = false
  }
}
</script>

<template>
  <p v-if="!liveDemo">Use SSH local port forwarding to reach the deployment from the presenting machine.</p>
  <section v-else class="demo-setup" @keydown.stop @keyup.stop>
    <div class="ssh-fields">
      <label>SSH host <input v-model.trim="ssh.host" placeholder="EC2 public DNS or IP" /></label>
      <label>SSH user <input v-model.trim="ssh.user" /></label>
      <label>SSH port <input v-model.number="ssh.port" type="number" min="1" max="65535" /></label>
      <label>Identity file path <input v-model.trim="ssh.keyPath" placeholder="Optional absolute path" /></label>
      <label>Local tunnel port <input v-model.number="ssh.localPort" type="number" min="1" max="65535" /></label>
      <label>HTTP address from SSH host <input v-model.trim="ssh.remoteHost" /></label>
      <label>HTTP port from SSH host <input v-model.number="ssh.remotePort" type="number" min="1" max="65535" /></label>
    </div>
    <pre v-if="configured" class="ssh-command">{{ command }}</pre>
    <p class="setup-help">Run the generated command in a terminal and leave it running. The HTTP port is the one reachable from the SSH host, typically Docker's published host port.</p>
    <div v-if="configured" class="demo-actions">
      <button type="button" @click="copyCommand">{{ copied ? 'Copied' : 'Copy SSH command' }}</button>
      <button type="button" @click="useTunnel">Use tunnel for deployment demo</button>
      <span class="setup-help">Deployment target {{ demo.deployment.host }}:{{ demo.deployment.port }}</span>
    </div>
  </section>
</template>
