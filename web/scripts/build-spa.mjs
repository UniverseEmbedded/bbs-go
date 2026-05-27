import { spawnSync } from "node:child_process"
import { existsSync } from "node:fs"
import path from "node:path"
import { fileURLToPath } from "node:url"

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..")
const reactRouterBin = path.join(
  root,
  "node_modules",
  ".bin",
  process.platform === "win32" ? "react-router.CMD" : "react-router"
)

const result = spawnSync(reactRouterBin, ["build"], {
  cwd: root,
  stdio: "inherit",
  env: {
    ...process.env,
    BBSGO_WEB_SPA: "true",
  },
  shell: process.platform === "win32",
})

if (result.error) {
  console.error(result.error)
  process.exit(1)
}
if (result.status !== 0) {
  process.exit(result.status ?? 1)
}

const clientIndex = path.join(root, "build", "client", "index.html")
if (!existsSync(clientIndex)) {
  console.error(`SPA build did not produce ${clientIndex}`)
  process.exit(1)
}

await import("./prepare-spa-build.mjs")

const spaIndex = path.join(root, "build", "spa", "index.html")
if (!existsSync(spaIndex)) {
  console.error(`SPA build did not produce ${spaIndex}`)
  process.exit(1)
}
