import assert from "node:assert/strict"
import { describe, it } from "node:test"
import ts from "typescript"
import { readFile } from "node:fs/promises"
import vm from "node:vm"

function plain(value) {
  return JSON.parse(JSON.stringify(value))
}

async function loadModule() {
  const source = await readFile(new URL("./workflow-utils.ts", import.meta.url), "utf8")
  const compiled = ts.transpileModule(source, {
    compilerOptions: {
      target: ts.ScriptTarget.ES2017,
      module: ts.ModuleKind.CommonJS,
    },
    fileName: "workflow-utils.ts",
  })
  const sandbox = {
    exports: {},
    module: { exports: {} },
  }
  sandbox.exports = sandbox.module.exports
  vm.runInNewContext(compiled.outputText, sandbox)
  return sandbox.module.exports
}

describe("validateWorkflowDraft", () => {
  it("rejects missing start", async () => {
    const { validateWorkflowDraft } = await loadModule()

    const result = validateWorkflowDraft({
      nodes: [{ id: "end_1", type: "end", position: { x: 0, y: 0 }, data: {} }],
      edges: [],
    })

    assert.equal(result.valid, false)
    assert.match(result.errors.join("\n"), /exactly one start/)
  })

  it("rejects dangling edge", async () => {
    const { validateWorkflowDraft } = await loadModule()

    const result = validateWorkflowDraft({
      nodes: [
        { id: "start_1", type: "start", position: { x: 0, y: 0 }, data: {} },
        { id: "end_1", type: "end", position: { x: 200, y: 0 }, data: {} },
      ],
      edges: [{ id: "e1", source: "start_1", target: "missing_1" }],
    })

    assert.equal(result.valid, false)
    assert.match(result.errors.join("\n"), /target node does not exist/)
  })
})

describe("toApiDefinition", () => {
  it("preserves xyflow node positions", async () => {
    const { toApiDefinition } = await loadModule()

    const definition = toApiDefinition({
      nodes: [
        {
          id: "start_1",
          type: "start",
          position: { x: 12, y: 34 },
          data: { name: "Start", config: { enabled: true } },
        },
        {
          id: "end_1",
          type: "end",
          position: { x: 240, y: 80 },
          data: { name: "End", config: {} },
        },
      ],
      edges: [{ id: "e1", source: "start_1", target: "end_1" }],
    })

    assert.deepEqual(plain(definition), {
      schemaVersion: 1,
      entryNodeId: "start_1",
      nodes: [
        {
          id: "start_1",
          type: "start",
          name: "Start",
          position: { x: 12, y: 34 },
          config: { enabled: true },
        },
        {
          id: "end_1",
          type: "end",
          name: "End",
          position: { x: 240, y: 80 },
          config: {},
        },
      ],
      edges: [{ id: "e1", source: "start_1", target: "end_1" }],
    })
  })

  it("uses node data type for xyflow default nodes", async () => {
    const { toApiDefinition } = await loadModule()

    const definition = toApiDefinition({
      nodes: [
        {
          id: "start_1",
          type: "default",
          position: { x: 0, y: 0 },
          data: { nodeType: "start", name: "Start", config: {} },
        },
        {
          id: "end_1",
          type: "default",
          position: { x: 200, y: 0 },
          data: { nodeType: "end", name: "End", config: {} },
        },
      ],
      edges: [{ id: "e1", source: "start_1", target: "end_1" }],
    })

    assert.equal(definition.entryNodeId, "start_1")
    assert.equal(definition.nodes[0].type, "start")
  })
})
