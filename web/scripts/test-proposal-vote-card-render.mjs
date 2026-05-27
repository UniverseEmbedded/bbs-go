import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"

const root = resolve(import.meta.dirname, "..")
const source = readFileSync(resolve(root, "components/topic/topic-vote-card.tsx"), "utf8")

assert.match(source, /function ProposalVoteCard\(/, "proposal vote card should exist")
assert.match(source, /\/api\/stance\/create/, "proposal vote card should submit stance")
assert.match(source, /\/api\/stance\/revoke\//, "proposal vote card should revoke stance")
assert.match(source, /pages\.topic\.detail\.vote\.proposalTag/, "proposal vote i18n key should be used")
assert.match(source, /canViewResults !== false/, "proposal card should respect result visibility")

console.log("proposal vote card structure OK")
