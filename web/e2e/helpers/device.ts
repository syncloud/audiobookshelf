import { ssh, scpTo } from './ssh'

const app = 'audiobookshelf'

export function installStoreVersion () {
  ssh(`snap remove ${app}`, { throw: false, timeout: 600_000 })
  ssh(`snap install ${app}`, { timeout: 1_200_000 })
}

export function inspect (label: string): string {
  return ssh(
    `echo "=== ${label} ==="; ` +
      `snap list ${app} 2>&1 || true; ` +
      `echo "svc: active=$(systemctl is-active snap.${app}.abs.service) NRestarts=$(systemctl show snap.${app}.abs.service -p NRestarts --value) since=$(systemctl show snap.${app}.abs.service -p ActiveEnterTimestamp --value)"; ` +
      `echo "--- config dir ---"; ls -la /var/snap/${app}/current/config 2>&1 || true; ` +
      `echo "--- installed marker ---"; ls -la /var/snap/${app}/common/installed 2>&1 || true; ` +
      `echo "--- configure-hook + restart events (snapd) ---"; journalctl -u snapd --no-pager 2>/dev/null | grep -iE "configure|restart|refresh|install" | tail -20 || true`,
    { throw: false }
  )
}

export function upgradeToBuild () {
  const snap = process.env.PLAYWRIGHT_SNAP
  if (!snap) throw new Error('PLAYWRIGHT_SNAP not set')
  scpTo(snap, `/tmp/${app}.snap`)
  ssh(`snap install --devmode /tmp/${app}.snap`, { timeout: 600_000 })
}
