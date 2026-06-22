"use client"

import "@xyflow/react/dist/style.css"

import {
  addEdge,
  Background,
  Controls,
  Handle,
  MiniMap,
  Position,
  ReactFlow,
  useEdgesState,
  useNodesState,
  type Connection,
  type Edge,
  type Node,
  type NodeProps,
} from "@xyflow/react"
import { AlertCircleIcon, CheckCircle2Icon, PlusIcon } from "lucide-react"
import { useCallback, useEffect, useMemo, useState } from "react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable"
import type { AIWorkflowDefinition, AIWorkflowNodeSpec } from "@/lib/api/admin"
import {
  applyAutoInputMappings,
  fromApiDefinition,
  getAvailableVariables,
  getNodeSpec,
  getRequiredInputs,
  toApiDefinition,
  validateWorkflowDraft,
  type WorkflowEditorEdge,
  type WorkflowEditorNode,
  type WorkflowNodeSpec,
} from "./workflow-utils"
import { NodeConfigPanel } from "./node-config-panel"

type WorkflowNodeData = Record<string, unknown> & {
  nodeType?: string
  name?: string
  config?: Record<string, unknown>
  inputs?: Record<string, { nodeId: string; field: string }>
  label?: string
  title?: string
  description?: string
  inputCount?: number
  outputCount?: number
  missingInputs?: string[]
}

type WorkflowFlowNode = Node<WorkflowNodeData>
type WorkflowFlowEdge = Edge

const nodeTypes = {
  workflowNode: WorkflowCanvasNode,
}

function toFlowNodes(definition: AIWorkflowDefinition): WorkflowFlowNode[] {
  return fromApiDefinition(definition).nodes.map((node) => ({
    id: node.id,
    type: "workflowNode",
    position: node.position,
    data: {
      nodeType: node.data?.nodeType ?? node.type,
      name: node.data?.name ?? node.id,
      label: node.data?.name ?? node.type ?? node.id,
      config: node.data?.config ?? {},
      inputs: node.data?.inputs ?? {},
    },
  }))
}

function toFlowEdges(definition: AIWorkflowDefinition): WorkflowFlowEdge[] {
  return (definition.edges ?? []).map((edge) => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    data: edge.condition ? { condition: edge.condition } : undefined,
  }))
}

function toDraft(nodes: WorkflowFlowNode[], edges: WorkflowFlowEdge[]) {
  return {
    nodes: nodes.map((node) => ({
      id: node.id,
      type: node.type,
      position: node.position,
      data: {
        nodeType: node.data.nodeType,
        name: node.data.name,
        config: node.data.config,
        inputs: node.data.inputs,
      },
    })) as WorkflowEditorNode[],
    edges: edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      data: edge.data as WorkflowEditorEdge["data"],
    })),
  }
}

export function WorkflowEditor({
  definition,
  nodeSpecs,
  onDefinitionChange,
}: {
  definition: AIWorkflowDefinition
  nodeSpecs: AIWorkflowNodeSpec[]
  onDefinitionChange: (definition: AIWorkflowDefinition) => void
}) {
  const [nodes, setNodes, onNodesChange] = useNodesState<WorkflowFlowNode>(
    toFlowNodes(definition)
  )
  const [edges, setEdges, onEdgesChange] = useEdgesState<WorkflowFlowEdge>(
    toFlowEdges(definition)
  )
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const selectedNode = useMemo(
    () => nodes.find((node) => node.id === selectedNodeId) ?? null,
    [nodes, selectedNodeId]
  )
  const draft = useMemo(() => toDraft(nodes, edges), [nodes, edges])
  const validation = useMemo(
    () => validateWorkflowDraft(draft, nodeSpecs),
    [draft, nodeSpecs]
  )
  const renderedNodes = useMemo(
    () => enrichNodesForRender(nodes, nodeSpecs),
    [nodes, nodeSpecs]
  )
  const selectedNodeSpec = useMemo(
    () => getNodeSpec(nodeSpecs, selectedNode?.data.nodeType ?? ""),
    [nodeSpecs, selectedNode]
  )
  const availableVariables = useMemo(
    () => (selectedNode ? getAvailableVariables(draft, selectedNode.id, nodeSpecs) : []),
    [draft, nodeSpecs, selectedNode]
  )

  useEffect(() => {
    onDefinitionChange(toApiDefinition(draft) as AIWorkflowDefinition)
  }, [draft, onDefinitionChange])

  const onConnect = useCallback(
    (connection: Connection) => {
      let newEdge: WorkflowFlowEdge | null = null
      setEdges((current) => {
        let nextIndex = current.length + 1
        let id = `edge_${connection.source}_${connection.target}_${nextIndex}`
        while (current.some((edge) => edge.id === id)) {
          nextIndex += 1
          id = `edge_${connection.source}_${connection.target}_${nextIndex}`
        }
        newEdge = {
          ...connection,
          id,
        } as WorkflowFlowEdge
        return addEdge(
          {
            ...connection,
            id,
          },
          current
        )
      })
      if (connection.source && connection.target) {
        setNodes((currentNodes) => {
          const currentDraft = toDraft(currentNodes, newEdge ? [...edges, newEdge] : edges)
          const nextDraft = applyAutoInputMappings(
            currentDraft,
            connection.source!,
            connection.target!,
            nodeSpecs
          )
          return currentNodes.map((node) => {
            const nextNode = nextDraft.nodes.find((item) => item.id === node.id)
            if (!nextNode) {
              return node
            }
            return {
              ...node,
              data: {
                ...node.data,
                inputs: nextNode.data?.inputs ?? node.data.inputs,
              },
            }
          })
        })
      }
    },
    [edges, nodeSpecs, setEdges, setNodes]
  )

  const addNode = (spec: AIWorkflowNodeSpec) => {
    setNodes((current) => {
      let nextIndex = current.length + 1
      let id = `${spec.type}_${nextIndex}`
      while (current.some((node) => node.id === id)) {
        nextIndex += 1
        id = `${spec.type}_${nextIndex}`
      }
      return [
        ...current,
        {
          id,
          type: "workflowNode",
          position: { x: 120 + current.length * 28, y: 100 + current.length * 24 },
          data: {
            nodeType: spec.type,
            name: spec.title,
            label: spec.title,
            config: {},
            inputs: spec.defaultInputs ?? {},
          },
        },
      ]
    })
  }

  const updateNodeData = (nodeId: string, data: WorkflowNodeData) => {
    setNodes((current) =>
      current.map((node) =>
        node.id === nodeId
          ? {
              ...node,
              data: {
                ...data,
                label: data.name ?? data.nodeType ?? node.id,
              },
            }
          : node
      )
    )
  }

  return (
    <ResizablePanelGroup orientation="horizontal" className="h-full min-h-0 border-t">
      <ResizablePanel defaultSize="18%" minSize="12%" maxSize="34%" className="min-h-0">
        <aside className="h-full min-h-0 overflow-y-auto bg-muted/20 p-3">
          <div className="mb-3 text-sm font-medium">节点库</div>
          <div className="space-y-2">
            {nodeSpecs.map((spec) => (
              <button
                key={spec.type}
                type="button"
                onClick={() => addNode(spec)}
                className="flex w-full items-start gap-2 rounded-md border bg-background px-3 py-2 text-left text-sm hover:bg-muted"
              >
                <PlusIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                <span className="min-w-0">
                  <span className="block truncate font-medium">{spec.title}</span>
                  <span className="mt-1 line-clamp-2 text-xs text-muted-foreground">
                    {spec.description}
                  </span>
                  <span className="mt-1 flex gap-2 text-[11px] text-muted-foreground">
                    <span>输入 {spec.inputSchema?.length ?? 0}</span>
                    <span>输出 {spec.outputSchema?.length ?? 0}</span>
                  </span>
                </span>
              </button>
            ))}
          </div>
        </aside>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="56%" minSize="30%" className="min-h-0">
        <section className="relative h-full min-h-0">
          <ReactFlow
            nodes={renderedNodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={(_, node) => setSelectedNodeId(node.id)}
            fitView
          >
            <Background />
            <Controls />
            <MiniMap pannable zoomable />
          </ReactFlow>
          <WorkflowValidationBadge errors={validation.errors} valid={validation.valid} />
        </section>
      </ResizablePanel>
      <ResizableHandle withHandle />
      <ResizablePanel defaultSize="26%" minSize="18%" maxSize="40%" className="min-h-0">
        <aside className="h-full min-h-0 overflow-y-auto bg-muted/10">
          <NodeConfigPanel
            node={selectedNode}
            nodeSpec={selectedNodeSpec}
            availableVariables={availableVariables}
            onChange={updateNodeData}
          />
          {!validation.valid ? (
            <div className="border-t p-4">
              <div className="mb-2 text-sm font-medium">Local validation</div>
              <ul className="space-y-1 text-xs text-destructive">
                {validation.errors.map((error) => (
                  <li key={error}>{error}</li>
                ))}
              </ul>
            </div>
          ) : null}
          <div className="border-t p-4">
            <Button
              variant="outline"
              className="w-full"
              onClick={() => onDefinitionChange(toApiDefinition(toDraft(nodes, edges)) as AIWorkflowDefinition)}
            >
              Sync definition
            </Button>
          </div>
        </aside>
      </ResizablePanel>
    </ResizablePanelGroup>
  )
}

function enrichNodesForRender(
  nodes: WorkflowFlowNode[],
  nodeSpecs: AIWorkflowNodeSpec[]
): WorkflowFlowNode[] {
  return nodes.map((node) => {
    const spec = getNodeSpec(nodeSpecs, node.data.nodeType ?? "")
    const missingInputs = getRequiredInputs(spec).filter((input) => {
      const selector = node.data.inputs?.[input.name]
      return !selector?.nodeId || !selector.field
    })
    return {
      ...node,
      data: {
        ...node.data,
        title: spec?.title ?? node.data.name ?? node.id,
        description: spec?.description ?? "",
        inputCount: spec?.inputSchema?.length ?? 0,
        outputCount: spec?.outputSchema?.length ?? 0,
        missingInputs: missingInputs.map((input) => input.name),
      },
    }
  })
}

function WorkflowCanvasNode({ data, selected }: NodeProps<WorkflowFlowNode>) {
  const missingInputs = data.missingInputs ?? []
  const hasIssue = missingInputs.length > 0
  return (
    <div
      className={[
        "min-w-56 rounded-md border bg-background shadow-sm",
        selected ? "ring-2 ring-ring" : "",
        hasIssue ? "border-destructive/70" : "border-border",
      ].join(" ")}
    >
      <Handle type="target" position={Position.Left} />
      <div className="flex items-start gap-2 border-b px-3 py-2">
        {hasIssue ? (
          <AlertCircleIcon className="mt-0.5 size-4 shrink-0 text-destructive" />
        ) : (
          <CheckCircle2Icon className="mt-0.5 size-4 shrink-0 text-emerald-600" />
        )}
        <div className="min-w-0 flex-1">
          <div className="truncate text-sm font-medium">{data.name ?? data.title}</div>
          <div className="mt-0.5 truncate text-xs text-muted-foreground">{data.title}</div>
        </div>
      </div>
      <div className="space-y-2 px-3 py-2 text-xs">
        <div className="flex justify-between text-muted-foreground">
          <span>输入 {data.inputCount ?? 0}</span>
          <span>输出 {data.outputCount ?? 0}</span>
        </div>
        {hasIssue ? (
          <div className="rounded-sm bg-destructive/10 px-2 py-1 text-destructive">
            缺少输入：{missingInputs.join("、")}
          </div>
        ) : (
          <div className="rounded-sm bg-emerald-500/10 px-2 py-1 text-emerald-700">
            配置完整
          </div>
        )}
      </div>
      <Handle type="source" position={Position.Right} />
    </div>
  )
}

function WorkflowValidationBadge({
  errors,
  valid,
}: {
  errors: string[]
  valid: boolean
}) {
  return (
    <div className="absolute left-3 top-3 flex gap-2">
      {valid ? (
        <Badge variant="default">Valid draft</Badge>
      ) : (
        <Popover>
          <PopoverTrigger
            render={
              <button
                type="button"
                className="inline-flex rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            }
          >
            <Badge variant="destructive" className="cursor-pointer">
              {errors.length} issues
            </Badge>
          </PopoverTrigger>
          <PopoverContent side="bottom" align="start" className="w-80">
            <div className="text-sm font-medium">Validation issues</div>
            <ul className="mt-2 max-h-72 space-y-1 overflow-y-auto text-xs text-destructive">
              {errors.map((error) => (
                <li key={error} className="rounded-md bg-destructive/10 px-2 py-1.5">
                  {error}
                </li>
              ))}
            </ul>
          </PopoverContent>
        </Popover>
      )}
    </div>
  )
}
