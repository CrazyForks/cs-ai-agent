export type WorkflowNodePosition = {
  x: number
  y: number
}

export type WorkflowEditorNode = {
  id: string
  type?: string
  position: WorkflowNodePosition
  data?: {
    nodeType?: string
    name?: string
    label?: string
    config?: Record<string, unknown>
    inputs?: Record<string, WorkflowVariableSelector>
  }
}

export type WorkflowEditorEdge = {
  id: string
  source: string
  target: string
  data?: {
    condition?: {
      expression?: string
      left?: WorkflowVariableSelector
      operator?: string
      right?: unknown
    }
  }
}

export type WorkflowDraft = {
  nodes: WorkflowEditorNode[]
  edges: WorkflowEditorEdge[]
}

export type WorkflowDefinition = {
  schemaVersion: number
  entryNodeId: string
  nodes: {
    id: string
    type: string
    name: string
    position: WorkflowNodePosition
    config: Record<string, unknown>
    inputs?: Record<string, WorkflowVariableSelector>
  }[]
  edges: {
    id: string
    source: string
    target: string
    condition?: {
      expression?: string
      left?: WorkflowVariableSelector
      operator?: string
      right?: unknown
    }
  }[]
}

export type WorkflowVariableType =
  | "string"
  | "integer"
  | "boolean"
  | "object"
  | "array<string>"
  | "array<int>"
  | "array<object>"
  | "any"

export type WorkflowVariableSelector = {
  nodeId: string
  field: string
}

export type WorkflowVariableSpec = {
  name: string
  type: WorkflowVariableType
  required?: boolean
  description?: string
}

export type WorkflowNodeSpec = {
  type: string
  title?: string
  description?: string
  inputSchema?: WorkflowVariableSpec[]
  outputSchema?: WorkflowVariableSpec[]
  defaultInputs?: Record<string, WorkflowVariableSelector>
}

export type WorkflowVariableRef = {
  nodeId: string
  nodeName: string
  field: string
  type: string
  description: string
}

export type WorkflowDraftValidation = {
  valid: boolean
  errors: string[]
}

export function validateWorkflowDraft(
  draft: WorkflowDraft,
  nodeSpecs: WorkflowNodeSpec[] = []
): WorkflowDraftValidation {
  const errors: string[] = []
  const nodeIds = new Set<string>()
  let startCount = 0
  let endCount = 0

  for (const node of draft.nodes) {
    const id = node.id.trim()
    if (!id) {
      errors.push("node id is required")
      continue
    }
    if (nodeIds.has(id)) {
      errors.push(`duplicate node id: ${id}`)
    }
    nodeIds.add(id)
    const nodeType = node.data?.nodeType ?? node.type
    if (nodeType === "start") {
      startCount += 1
    }
    if (nodeType === "end") {
      endCount += 1
    }
  }

  if (startCount !== 1) {
    errors.push("workflow must contain exactly one start node")
  }
  if (endCount < 1) {
    errors.push("workflow must contain at least one end node")
  }

  const edgeIds = new Set<string>()
  const conditionalSources = new Set<string>()
  const defaultSources = new Set<string>()
  for (const edge of draft.edges) {
    const id = edge.id.trim()
    if (!id) {
      errors.push("edge id is required")
    } else if (edgeIds.has(id)) {
      errors.push(`duplicate edge id: ${id}`)
    }
    edgeIds.add(id)
    if (!nodeIds.has(edge.source)) {
      errors.push(`edge source node does not exist: ${edge.source}`)
    }
    if (!nodeIds.has(edge.target)) {
      errors.push(`edge target node does not exist: ${edge.target}`)
    }
    if (edge.data?.condition) {
      conditionalSources.add(edge.source)
      if (!edge.data.condition.left?.nodeId || !edge.data.condition.left.field) {
        errors.push(`edge ${edge.id} condition left variable is required`)
      }
      if (!edge.data.condition.operator) {
        errors.push(`edge ${edge.id} condition operator is required`)
      }
    } else {
      defaultSources.add(edge.source)
    }
  }
  for (const source of conditionalSources) {
    if (!defaultSources.has(source)) {
      errors.push(`node ${source} conditional branch must include a default edge`)
    }
  }

  for (const node of draft.nodes) {
    const nodeType = node.data?.nodeType ?? node.type ?? ""
    const spec = getNodeSpec(nodeSpecs, nodeType)
    if (!spec) {
      continue
    }
    for (const input of getRequiredInputs(spec)) {
      const selector = node.data?.inputs?.[input.name]
      if (!selector?.nodeId || !selector.field) {
        const nodeName = node.data?.name ?? spec.title ?? node.id
        errors.push(`${nodeName} 缺少必填输入「${input.name}」，请选择上游节点的输出变量。`)
      }
    }
  }

  return {
    valid: errors.length === 0,
    errors,
  }
}

export function toApiDefinition(draft: WorkflowDraft): WorkflowDefinition {
  const startNode = draft.nodes.find((node) => (node.data?.nodeType ?? node.type) === "start")
  return {
    schemaVersion: 1,
    entryNodeId: startNode?.id ?? "",
    nodes: draft.nodes.map((node) => ({
      id: node.id,
      type: node.data?.nodeType ?? node.type ?? "",
      name: node.data?.name ?? node.type ?? node.id,
      position: {
        x: node.position.x,
        y: node.position.y,
      },
      config: node.data?.config ?? {},
      ...(node.data?.inputs ? { inputs: node.data.inputs } : {}),
    })),
    edges: draft.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      ...(edge.data?.condition
        ? {
            condition: {
              ...(edge.data.condition.expression ? { expression: edge.data.condition.expression } : {}),
              ...(edge.data.condition.left ? { left: edge.data.condition.left } : {}),
              ...(edge.data.condition.operator ? { operator: edge.data.condition.operator } : {}),
              ...(edge.data.condition.right !== undefined ? { right: edge.data.condition.right } : {}),
            },
          }
        : {}),
    })),
  }
}

export function fromApiDefinition(definition: WorkflowDefinition): WorkflowDraft {
  return {
    nodes: (definition.nodes ?? []).map((node) => ({
      id: node.id,
      type: node.type,
      position: node.position ?? { x: 0, y: 0 },
      data: {
        nodeType: node.type,
        name: node.name,
        config: node.config ?? {},
        inputs: node.inputs ?? {},
      },
    })),
    edges: (definition.edges ?? []).map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      data: edge.condition ? { condition: edge.condition } : undefined,
    })),
  }
}

export function applyAutoInputMappings(
  draft: WorkflowDraft,
  sourceNodeId: string,
  targetNodeId: string,
  nodeSpecs: WorkflowNodeSpec[]
): WorkflowDraft {
  const sourceNode = draft.nodes.find((node) => node.id === sourceNodeId)
  const targetNode = draft.nodes.find((node) => node.id === targetNodeId)
  if (!sourceNode || !targetNode) {
    return draft
  }
  const sourceSpec = getNodeSpec(nodeSpecs, sourceNode.data?.nodeType ?? sourceNode.type ?? "")
  const targetSpec = getNodeSpec(nodeSpecs, targetNode.data?.nodeType ?? targetNode.type ?? "")
  if (!sourceSpec || !targetSpec) {
    return draft
  }
  const nextInputs = { ...(targetNode.data?.inputs ?? {}) }
  let changed = false

  for (const input of targetSpec.inputSchema ?? []) {
    if (nextInputs[input.name]) {
      continue
    }
    const output = findPreferredOutput(input.name, input.type, sourceSpec.outputSchema ?? [])
    if (!output) {
      continue
    }
    nextInputs[input.name] = { nodeId: sourceNodeId, field: output.name }
    changed = true
  }

  if (!changed) {
    return draft
  }

  return {
    ...draft,
    nodes: draft.nodes.map((node) =>
      node.id === targetNodeId
        ? {
            ...node,
            data: {
              ...node.data,
              inputs: nextInputs,
            },
          }
        : node
    ),
  }
}

export function createWorkflowNodeFromSpec(
  spec: WorkflowNodeSpec,
  existingNodes: Pick<WorkflowEditorNode, "id">[],
  position: WorkflowNodePosition
): WorkflowEditorNode {
  const id = uniqueNodeId(existingNodes, spec.type)
  return {
    id,
    type: "workflowNode",
    position,
    data: {
      nodeType: spec.type,
      name: spec.title ?? spec.type,
      label: spec.title ?? spec.type,
      config: {},
      inputs: spec.defaultInputs ?? {},
    },
  }
}

function uniqueNodeId(existingNodes: Pick<WorkflowEditorNode, "id">[], nodeType: string) {
  let nextIndex = existingNodes.length + 1
  let id = `${nodeType}_${nextIndex}`
  while (existingNodes.some((node) => node.id === id)) {
    nextIndex += 1
    id = `${nodeType}_${nextIndex}`
  }
  return id
}

function findPreferredOutput(
  inputName: string,
  inputType: WorkflowVariableType,
  outputs: WorkflowVariableSpec[]
): WorkflowVariableSpec | undefined {
  const preferred = preferredOutputName(inputName)
  if (preferred) {
    const exact = outputs.find((output) => output.name === preferred && variableTypesCompatible(inputType, output.type))
    if (exact) {
      return exact
    }
  }
  const sameName = outputs.find((output) => output.name === inputName && variableTypesCompatible(inputType, output.type))
  if (sameName) {
    return sameName
  }
  return outputs.find((output) => variableTypesCompatible(inputType, output.type))
}

function preferredOutputName(inputName: string): string {
  switch (inputName) {
    case "query":
    case "userMessage":
    case "issue":
    case "prompt":
      return "userMessage"
    case "knowledgeItems":
      return "items"
    case "replyText":
      return "replyText"
    case "confirmed":
      return "confirmed"
    case "ticketDraft":
      return "ticketDraft"
    case "reason":
      return "reason"
    default:
      return ""
  }
}

function variableTypesCompatible(input: WorkflowVariableType, output: WorkflowVariableType): boolean {
  return input === "any" || output === "any" || input === output
}

export function getNodeSpec(
  nodeSpecs: WorkflowNodeSpec[],
  nodeType: string
): WorkflowNodeSpec | undefined {
  return nodeSpecs.find((spec) => spec.type === nodeType)
}

export function getRequiredInputs(spec: WorkflowNodeSpec | undefined): WorkflowVariableSpec[] {
  return (spec?.inputSchema ?? []).filter((item) => item.required)
}

export function getAvailableVariables(
  draft: WorkflowDraft,
  nodeId: string,
  nodeSpecs: WorkflowNodeSpec[]
): WorkflowVariableRef[] {
  const ancestors = collectAncestorNodeIds(draft, nodeId)
  const nodesById = new Map(draft.nodes.map((node) => [node.id, node]))
  const variables: WorkflowVariableRef[] = []

  for (const sourceNodeId of ancestors) {
    const sourceNode = nodesById.get(sourceNodeId)
    if (!sourceNode) {
      continue
    }
    const nodeType = sourceNode.data?.nodeType ?? sourceNode.type ?? ""
    const spec = getNodeSpec(nodeSpecs, nodeType)
    for (const output of spec?.outputSchema ?? []) {
      variables.push({
        nodeId: sourceNode.id,
        nodeName: sourceNode.data?.name ?? spec?.title ?? sourceNode.id,
        field: output.name,
        type: output.type,
        description: output.description ?? "",
      })
    }
  }

  return variables
}

function collectAncestorNodeIds(draft: WorkflowDraft, nodeId: string): string[] {
  const incoming = new Map<string, string[]>()
  for (const edge of draft.edges) {
    const sources = incoming.get(edge.target) ?? []
    sources.push(edge.source)
    incoming.set(edge.target, sources)
  }

  const visited = new Set<string>()
  const ordered: string[] = []

  function visit(current: string) {
    for (const source of incoming.get(current) ?? []) {
      if (visited.has(source)) {
        continue
      }
      visited.add(source)
      visit(source)
      ordered.push(source)
    }
  }

  visit(nodeId)
  return ordered
}
