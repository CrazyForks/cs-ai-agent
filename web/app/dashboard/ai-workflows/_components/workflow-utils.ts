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
    config?: Record<string, unknown>
  }
}

export type WorkflowEditorEdge = {
  id: string
  source: string
  target: string
  data?: {
    condition?: {
      expression: string
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
  }[]
  edges: {
    id: string
    source: string
    target: string
    condition?: {
      expression: string
    }
  }[]
}

export type WorkflowDraftValidation = {
  valid: boolean
  errors: string[]
}

export function validateWorkflowDraft(draft: WorkflowDraft): WorkflowDraftValidation {
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
    })),
    edges: draft.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      ...(edge.data?.condition
        ? {
            condition: {
              expression: edge.data.condition.expression,
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
