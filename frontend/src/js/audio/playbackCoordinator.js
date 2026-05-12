let currentStop = null

export function claimPlayback(stopFn) {
  if (currentStop && currentStop !== stopFn) {
    currentStop()
  }
  currentStop = stopFn
}

export function releasePlayback(stopFn) {
  if (currentStop === stopFn) {
    currentStop = null
  }
}

export function stopPlayback() {
  currentStop?.()
  currentStop = null
}
