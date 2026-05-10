class PCMPlayerProcessor extends AudioWorkletProcessor {
  constructor() {
    super()
    this.queue = []
    this.current = null
    this.offset = 0
    this.maxQueuedSamples = 24000 * 3
    this.queuedSamples = 0

    this.port.onmessage = event => {
      if (event.data?.type === 'reset') {
        this.queue = []
        this.current = null
        this.offset = 0
        this.queuedSamples = 0
        return
      }

      if (event.data?.type !== 'pcm' || !event.data.samples) {
        return
      }

      const samples = new Float32Array(event.data.samples)
      this.queue.push(samples)
      this.queuedSamples += samples.length

      while (this.queuedSamples > this.maxQueuedSamples && this.queue.length > 1) {
        this.queuedSamples -= this.queue.shift().length
      }
    }
  }

  process(inputs, outputs) {
    const output = outputs[0][0]

    for (let i = 0; i < output.length; i++) {
      if (!this.current || this.offset >= this.current.length) {
        this.current = this.queue.shift() || null
        this.offset = 0
        if (this.current) {
          this.queuedSamples -= this.current.length
        }
      }

      output[i] = this.current ? this.current[this.offset++] : 0
    }

    return true
  }
}

registerProcessor('pcm-player', PCMPlayerProcessor)
