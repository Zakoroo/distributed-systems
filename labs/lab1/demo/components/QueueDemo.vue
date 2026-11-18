<script setup>
import { reactive, ref } from 'vue'

const capacity = 3
const messages = ['Rain', 'Clear skies', 'Wind', 'Fog', 'Sun']
const subscribers = reactive([
  { name: 'Alice', queue: [], update: '', skipped: false },
  { name: 'Bob', queue: [], update: '', skipped: false },
])
const sequence = ref(0)
const summary = ref('Publish an event to put a copy in each subscriber queue.')

function cellsFor(queue) {
  // Give events stable keys so moving toward the front animates the same card.
  return Array.from({ length: capacity }, (_, index) => {
    const event = queue[index]
    return { key: event?.sequence ?? 'empty-' + index, event }
  })
}

function publish() {
  const event = {
    sequence: ++sequence.value,
    text: messages[(sequence.value - 1) % messages.length],
  }

  for (const subscriber of subscribers) {
    subscriber.skipped = subscriber.queue.length >= capacity
    if (subscriber.skipped) {
      subscriber.update = 'Skipped #' + event.sequence + ' · queue full'
    } else {
      subscriber.queue.push(event)
      subscriber.update = 'Added #' + event.sequence
    }
  }
  summary.value = 'Published #' + event.sequence + ' · ' + event.text
}

function read(subscriber) {
  const event = subscriber.queue.shift()
  if (!event) return

  for (const entry of subscribers) {
    entry.skipped = false
    entry.update = entry === subscriber ? 'Read #' + event.sequence : 'Queue stayed the same'
  }
  summary.value = subscriber.name + ' read #' + event.sequence + ' · ' + event.text
}

function reset() {
  sequence.value = 0
  for (const subscriber of subscribers) {
    subscriber.queue = []
    subscriber.update = ''
    subscriber.skipped = false
  }
  summary.value = 'Publish an event to put a copy in each subscriber queue.'
}
</script>

<template>
  <section class="queue-demo" @keydown.stop @keyup.stop>
    <div class="queue-toolbar">
      <div class="queue-topic">Topic <code>weather</code></div>
      <div class="queue-controls">
        <button type="button" class="queue-publish" @click="publish">Publish event</button>
        <button type="button" @click="reset">Reset</button>
      </div>
    </div>

    <div class="queue-summary" aria-live="polite">{{ summary }}</div>

    <div class="queue-lanes">
      <article v-for="subscriber in subscribers" :key="subscriber.name" class="queue-lane">
        <div class="queue-lane-heading">
          <strong>{{ subscriber.name }}</strong>
          <span>{{ subscriber.queue.length }} / {{ capacity }} queued</span>
        </div>

        <div class="queue-direction"><span>Oldest</span><span aria-hidden="true">→</span><span>Newest</span></div>
        <TransitionGroup name="queue-event" tag="ol" class="queue-cells" :aria-label="subscriber.name + ' queue, oldest first'">
          <li
            v-for="({ key, event }, index) in cellsFor(subscriber.queue)"
            :key="key"
            class="queue-cell"
            :class="{ occupied: event, oldest: index === 0 && event }"
          >
            <template v-if="event">
              <code>#{{ event.sequence }}</code>
              <span>{{ event.text }}</span>
            </template>
            <span v-else class="queue-empty">Empty</span>
          </li>
        </TransitionGroup>

        <div class="queue-lane-footer">
          <button type="button" :disabled="!subscriber.queue.length" @click="read(subscriber)">
            Read next as {{ subscriber.name }}
          </button>
          <span class="queue-update" :class="{ skipped: subscriber.skipped }" aria-live="polite">
            {{ subscriber.update || 'Ready for events' }}
          </span>
        </div>
      </article>
    </div>

    <p class="queue-caption">This example uses 3 places per queue. The live broker uses 64 and skips new events for a subscriber whose queue is full.</p>
  </section>
</template>

<style scoped>
.queue-toolbar, .queue-controls, .queue-lane-heading, .queue-lane-footer, .queue-direction {
  display: flex;
  align-items: center;
  gap: 10px;
}
.queue-toolbar, .queue-lane-heading, .queue-lane-footer, .queue-direction { justify-content: space-between; }
.queue-toolbar {
  background: var(--deck-panel);
  border: 1px solid var(--deck-line);
  border-radius: 12px;
  padding: 10px 14px;
}
.queue-topic { font-size: 1rem; font-weight: 600; }
.queue-demo button {
  padding: 5px 10px;
  border: 1px solid var(--deck-line);
  border-radius: 6px;
  color: var(--deck-ink);
  background: var(--deck-surface);
  font-size: .78rem;
  font-weight: 650;
}
.queue-demo button.queue-publish { background: var(--deck-button); color: white; border-color: transparent; }
.queue-demo button:disabled { opacity: .5; cursor: default; }
.queue-demo button:focus-visible { outline: 2px solid var(--deck-accent); outline-offset: 2px; }
.queue-summary {
  margin: 12px 0;
  padding-left: 10px;
  border-left: 3px solid var(--deck-accent);
  font-size: .92rem;
  font-weight: 650;
}
.queue-lanes { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 14px; }
.queue-lane {
  padding: 12px 14px;
  border: 1px solid var(--deck-line);
  border-radius: 12px;
  background: var(--deck-surface);
}
.queue-lane-heading { font-size: 1rem; }
.queue-lane-heading > span, .queue-direction, .queue-update { font-size: .75rem; color: var(--deck-link); }
.queue-direction { margin-top: 10px; }
.queue-direction > span:nth-child(2) { flex: 1; text-align: center; }
ol.queue-cells {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  list-style: none;
  padding: 0;
  margin: 5px 0 12px;
}
.queue-cells .queue-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  min-height: 60px;
  margin: 0;
  padding: 4px;
  border: 1px dashed var(--deck-line);
  border-radius: 7px;
  font-size: .78rem;
}
.queue-cell.occupied { background: var(--deck-panel); border-style: solid; }
.queue-event-move { transition: transform var(--deck-reveal-time) var(--deck-reveal-ease); }
.queue-event-leave-active { position: absolute; visibility: hidden; transition: none !important; }
.queue-event-enter-active.occupied {
  transition: opacity var(--deck-reveal-time) ease, translate var(--deck-reveal-time) var(--deck-reveal-ease);
}
.queue-event-enter-from.occupied { opacity: 0; translate: 6px 0; }
@media (prefers-reduced-motion: reduce) {
  .queue-event-enter-from.occupied { translate: none; }
}
.queue-cell.oldest { border-color: var(--deck-accent); }
.queue-cell code { color: var(--deck-link); font-weight: 750; }
.queue-empty { color: var(--deck-link); opacity: .8; }
.queue-update { text-align: right; }
.queue-update.skipped { color: var(--deck-warm); font-weight: 700; }
.queue-demo .queue-caption { margin: 12px 0 0; font-size: .8rem; line-height: 1.35; }
</style>
