import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { resolve } from "node:path"

const root = resolve(import.meta.dirname, "..")
const source = readFileSync(resolve(root, "components/topic/topic-create-form.tsx"), "utf8")

for (const key of [
  "pages.topic.create.vote.modePoll",
  "pages.topic.create.vote.modeProposal",
  "pages.topic.create.vote.proposalSettings",
  "pages.topic.create.vote.hideResultsOff",
  "pages.topic.create.vote.reasonDisabled",
  "pages.topic.create.vote.meaningPlaceholder",
  "pages.topic.create.vote.validateReasonMode",
]) {
  assert.ok(source.includes(key), `expected create form to use ${key}`)
}

assert.match(source, /pollType: \"proposal\"/, "proposal mode should be present in create form")
assert.match(source, /meaning:\s*option\.meaning\?\.trim\(\) \|\| \"\"/, "proposal meaning should be sent")
assert.match(source, /prompt:\s*option\.prompt\?\.trim\(\) \|\| \"\"/, "proposal prompt should be sent")

console.log("proposal topic create form structure OK")
