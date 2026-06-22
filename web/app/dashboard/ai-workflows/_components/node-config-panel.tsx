"use client"

import { useState } from "react"
import type { Node } from "@xyflow/react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { VariableSelector } from "./variable-selector"
import type {
  WorkflowNodeSpec,
  WorkflowVariableRef,
  WorkflowVariableSelector,
} from "./workflow-utils"

type WorkflowNodeData = Record<string, unknown> & {
  nodeType?: string
  name?: string
  config?: Record<string, unknown>
  inputs?: Record<string, WorkflowVariableSelector>
}

export function NodeConfigPanel({
  node,
  nodeSpec,
  availableVariables,
  onChange,
}: {
  node: Node<WorkflowNodeData> | null
  nodeSpec?: WorkflowNodeSpec
  availableVariables: WorkflowVariableRef[]
  onChange: (nodeId: string, data: WorkflowNodeData) => void
}) {
  if (!node) {
    return (
      <div className="flex h-full items-center justify-center px-4 text-sm text-muted-foreground">
        选择一个节点后，可以配置输入映射并查看输出变量。
      </div>
    )
  }

  return (
    <NodeConfigForm
      key={node.id}
      node={node}
      nodeSpec={nodeSpec}
      availableVariables={availableVariables}
      onChange={onChange}
    />
  )
}

function NodeConfigForm({
  node,
  nodeSpec,
  availableVariables,
  onChange,
}: {
  node: Node<WorkflowNodeData>
  nodeSpec?: WorkflowNodeSpec
  availableVariables: WorkflowVariableRef[]
  onChange: (nodeId: string, data: WorkflowNodeData) => void
}) {
  const [name, setName] = useState(node.data.name ?? "")
  const [configText, setConfigText] = useState(JSON.stringify(node.data.config ?? {}, null, 2))
  const [inputs, setInputs] = useState<Record<string, WorkflowVariableSelector>>(
    node.data.inputs ?? {}
  )
  const [error, setError] = useState("")
  const inputSchema = nodeSpec?.inputSchema ?? []
  const outputSchema = nodeSpec?.outputSchema ?? []

  const commitChange = (next: Partial<WorkflowNodeData>) => {
    onChange(node.id, {
      ...node.data,
      name: name.trim() || node.data.nodeType || node.id,
      config: node.data.config ?? {},
      inputs,
      ...next,
    })
  }

  const handleApply = () => {
    try {
      const parsed = JSON.parse(configText || "{}") as Record<string, unknown>
      setError("")
      commitChange({ config: parsed })
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
        <Label htmlFor="workflow-node-name">节点名称</Label>
        <Input
          id="workflow-node-name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          onBlur={() => commitChange({ name: name.trim() || node.data.nodeType || node.id })}
        />
      </div>
      {inputSchema.length > 0 ? (
        <div className="space-y-3">
          <div className="text-sm font-medium">输入映射</div>
          {availableVariables.length === 0 ? (
            <div className="rounded-md border border-dashed p-2 text-xs text-muted-foreground">
              当前节点前面还没有可用变量，请先连接上游节点。
            </div>
          ) : null}
          {inputSchema.map((input) => (
            <div key={input.name} className="space-y-1.5">
              <div className="flex items-center justify-between gap-2">
                <Label className="text-xs">
                  {input.name}
                  {input.required ? <span className="text-destructive"> *</span> : null}
                </Label>
                <span className="text-xs text-muted-foreground">{input.type}</span>
              </div>
              <VariableSelector
                value={inputs[input.name]}
                variables={availableVariables}
                onChange={(value) => {
                  const nextInputs = {
                    ...inputs,
                    [input.name]: value,
                  }
                  setInputs(nextInputs)
                  commitChange({
                    inputs: nextInputs,
                  })
                }}
              />
              {inputs[input.name] ? (
                <div className="text-xs text-muted-foreground">
                  已选择：{inputs[input.name].nodeId}.{inputs[input.name].field}
                </div>
              ) : null}
              {input.description ? (
                <div className="text-xs text-muted-foreground">{input.description}</div>
              ) : null}
            </div>
          ))}
        </div>
      ) : null}
      <details className="rounded-md border bg-background p-3">
        <summary className="cursor-pointer text-sm font-medium">高级配置 JSON</summary>
        <div className="mt-3 space-y-2">
          <Textarea
            id="workflow-node-config"
            className="h-40 font-mono text-xs"
            value={configText}
            onChange={(event) => setConfigText(event.target.value)}
          />
          {error ? <div className="text-xs text-destructive">{error}</div> : null}
          <Button type="button" variant="outline" size="sm" onClick={handleApply}>
            保存高级配置
          </Button>
        </div>
      </details>
      {outputSchema.length > 0 ? (
        <div className="space-y-2">
          <div className="text-sm font-medium">输出变量</div>
          <div className="space-y-1 rounded-md border bg-background p-2">
            {outputSchema.map((output) => (
              <div key={output.name} className="space-y-0.5 rounded-sm px-1 py-0.5">
                <div className="flex items-center justify-between gap-2 text-xs">
                  <span className="truncate font-medium">{output.name}</span>
                  <span className="shrink-0 text-muted-foreground">{output.type}</span>
                </div>
                {output.description ? (
                  <div className="text-xs text-muted-foreground">{output.description}</div>
                ) : null}
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </div>
  )
}
