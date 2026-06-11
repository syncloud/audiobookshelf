import { ssh, scpTo } from './ssh'

const app = 'audiobookshelf'
const service = `snap.${app}.abs.service`

export function waitForStable (minSeconds = 90) {
  ssh(
    `for i in $(seq 1 300); do ` +
      `s=$(systemctl is-active ${service} 2>/dev/null); ` +
      `t=$(systemctl show ${service} -p ActiveEnterTimestampMonotonic --value 2>/dev/null); ` +
      `up=$(awk '{print int($1)}' /proc/uptime); ` +
      `u=$(( up - t / 1000000 )); ` +
      `if [ "$s" = active ] && [ "$t" -gt 0 ] && [ "$u" -ge ${minSeconds} ]; then echo "stable for $u seconds"; exit 0; fi; ` +
      `sleep 5; ` +
      `done; echo "service did not stabilize"; exit 1`,
    { timeout: 1_700_000 }
  )
}

export function installStoreVersion () {
  ssh(`snap remove ${app}`, { throw: false, timeout: 600_000 })
  ssh(`snap install ${app}`, { timeout: 1_200_000 })
  waitForStable()
}

export function upgradeToBuild () {
  const snap = process.env.PLAYWRIGHT_SNAP
  if (!snap) throw new Error('PLAYWRIGHT_SNAP not set')
  scpTo(snap, `/tmp/${app}.snap`)
  ssh(`snap install --dangerous /tmp/${app}.snap`, { timeout: 600_000 })
  waitForStable()
}
