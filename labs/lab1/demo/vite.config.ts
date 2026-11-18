import { request as httpRequest } from 'node:http'
import { request as httpsRequest } from 'node:https'
import { isIP } from 'node:net'
import { defineConfig } from 'vite'

export default defineConfig({
  server: { host: 'localhost' },
  plugins: [{
    name: 'lab-demo-client',
    configureServer(server) {
      server.middlewares.use('/lab-demo/request', (req, res) => {
        function fail(status: number, message: string) {
          if (res.headersSent) {
            res.destroy()
            return
          }
          res.writeHead(status, { 'Content-Type': 'text/plain; charset=utf-8', 'X-Lab-Demo-Error': '1' })
          res.end(message)
        }

        // Only the local deck can ask this development server to open a connection.
        const host = req.headers.host ?? ''
        const loopback = ['127.0.0.1', '::1', '::ffff:127.0.0.1'].includes(req.socket.remoteAddress ?? '')
        if (!loopback || !/^(localhost|127\.0\.0\.1|\[::1\])(:\d+)?$/.test(host)
          || (req.headers.origin && req.headers.origin !== 'http://' + host)
          || req.headers['x-lab-demo'] !== '1') {
          fail(403, 'Open the demo on localhost and send requests from the live panel.')
          return
        }
        if (req.method !== 'GET' && req.method !== 'POST') {
          fail(405, 'The demo sends GET and POST requests.')
          return
        }

        let target: URL
        const address = String(req.headers['x-lab-source-address'] ?? '')
        const port = Number(req.headers['x-lab-source-port'] ?? 0)
        try {
          const origin = new URL(String(req.headers['x-lab-origin']))
          const path = String(req.headers['x-lab-path'])
          if (!['http:', 'https:'].includes(origin.protocol) || origin.username || origin.password
            || origin.pathname !== '/' || origin.search || origin.hash || !path.startsWith('/')) {
            throw new Error()
          }
          target = new URL(origin.origin + path)
        } catch {
          fail(400, 'Enter an HTTP or HTTPS server address and a valid request path.')
          return
        }
        if (!isIP(address) || !Number.isInteger(port) || port < 0 || port > 65535) {
          fail(400, 'Use a local client IP and a source port between 0 and 65535.')
          return
        }

        const send = target.protocol === 'https:' ? httpsRequest : httpRequest
        const headers = { Connection: 'close' } as Record<string, string>
        if (req.method === 'POST') {
          headers['Content-Type'] = req.headers['content-type'] ?? 'application/octet-stream'
          if (req.headers['content-length']) headers['Content-Length'] = req.headers['content-length']
        }
        const outgoing = send(target, {
          method: req.method,
          localAddress: address,
          localPort: port,
          agent: false,
          headers,
          signal: AbortSignal.timeout(22000),
        }, upstream => {
          res.writeHead(upstream.statusCode ?? 502, {
            'Content-Type': 'application/octet-stream',
            'X-Lab-Source-Address': outgoing.socket?.localAddress ?? address,
            'X-Lab-Source-Port': String(outgoing.socket?.localPort ?? port),
          })
          upstream.on('error', () => res.destroy())
          upstream.pipe(res)
        })

        outgoing.on('error', (error: NodeJS.ErrnoException) => {
          const hint = error.code === 'EADDRINUSE'
            ? ' The source port is busy. Pick another client or use an automatic port.'
            : error.code === 'EADDRNOTAVAIL'
              ? ' Check that the IP is assigned to this machine and matches the server address family.'
              : ''
          fail(502, error.message + hint)
        })
        res.on('close', () => { if (!res.writableFinished) outgoing.destroy() })
        if (req.method === 'POST') req.pipe(outgoing)
        else outgoing.end()
      })
    },
  }],
})
