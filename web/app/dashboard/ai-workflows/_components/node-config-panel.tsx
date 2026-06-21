"use client"

import { useState } from "react"
import type { Node } from "@xyflow/react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"

type WorkflowNodeData = Record<string, unknown> & {
  nodeType?: string
  name?: string
  config?: Record<string, unknown>
}

export function NodeConfigPanel({
  node,
  onChange,
}: {
  node: Node<WorkflowNodeData> | null
  onChange: (nodeId: string, data: WorkflowNodeData) => void
}) {
  if (!node) {
    return (
      <div className="flex h-full items-center justify-center px-4 text-sm text-muted-foreground">
        Select a node to edit its properties.
      </div>
    )
  }

  return <NodeConfigForm key={node.id} node={node} onChange={onChange} />
}

function NodeConfigForm({
  node,
  onChange,
}: {
  node: Node<WorkflowNodeData>
  onChange: (nodeId: string, data: WorkflowNodeData) => void
}) {
  const [name, setName] = useState(node.data.name ?? "")
  const [configText, setConfigText] = useState(JSON.stringify(node.data.config ?? {}, null, 2))
  const [error, setError] = useState("")

  const handleApply = () => {
    try {
      const parsed = JSON.parse(configText || "{}") as Record<string, unknown>
      setError("")
      onChange(node.id, {
        ...node.data,
        name: name.trim() || node.data.nodeType || node.id,
        config: parsed,
      })
    } catch {
      setError("Config must be valid JSON.")
    }
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 p-4">
      <div>
        <div className="text-sm font-medium">{node.data.nodeType ?? node.id}</div>
        <div className="mt-1 text-xs text-muted-foreground">{node.id}</div>
      </div>
      <div className="space-y-2">
        <Label htmlFor="workflow-node-name">Name</Label>
        <Input
          id="workflow-node-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
        />
      </div>
      <div className="min-h-0 flex-1 space-y-2">
        <Label htmlFor="workflow-node-config">Config JSON</Label>
        <Textarea
          id="workflow-node-config"
          className="h-64 font-mono text-xs"
          value={configText}
          onChange={(event) => setConfigText(event.target.value)}
        />
      </div>
      {error ? <div className="text-xs text-destructive">{error}</div> : null}
      <Button onClick={handleApply}>Apply</Button>
    </div>
  )
}
