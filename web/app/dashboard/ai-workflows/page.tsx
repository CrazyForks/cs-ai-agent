"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { CheckCircle2Icon, GitBranchIcon, SaveIcon, SendIcon } from "lucide-react"
import { toast } from "sonner"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  createAIWorkflow,
  fetchAIWorkflowNodeSpecs,
  fetchAIWorkflows,
  publishAIWorkflow,
  updateAIWorkflow,
  validateAIWorkflow,
  type AIWorkflow,
  type AIWorkflowDefinition,
  type AIWorkflowNodeSpec,
  type AIWorkflowValidationResult,
} from "@/lib/api/admin"
import { WorkflowEditor } from "./_components/workflow-editor"

const emptyDefinition: AIWorkflowDefinition = {
  schemaVersion: 1,
  entryNodeId: "start_1",
  nodes: [
    {
      id: "start_1",
      type: "start",
      name: "Start",
      position: { x: 0, y: 80 },
      config: {},
    },
    {
      id: "end_1",
      type: "end",
      name: "End",
      position: { x: 360, y: 80 },
      config: {},
    },
  ],
  edges: [{ id: "edge_start_end", source: "start_1", target: "end_1" }],
}

export default function DashboardAIWorkflowsPage() {
  const [workflows, setWorkflows] = useState<AIWorkflow[]>([])
  const [nodeSpecs, setNodeSpecs] = useState<AIWorkflowNodeSpec[]>([])
  const [selected, setSelected] = useState<AIWorkflow | null>(null)
  const [name, setName] = useState("Customer support flow")
  const [description, setDescription] = useState("")
  const [ownerId, setOwnerId] = useState("1")
  const [definition, setDefinition] = useState<AIWorkflowDefinition>(emptyDefinition)
  const [validation, setValidation] = useState<AIWorkflowValidationResult | null>(null)
  const [loading, setLoading] = useState(false)
  const editorKey = useMemo(
    () => `${selected?.id ?? "new"}-${selected?.updatedAt ?? ""}`,
    [selected?.id, selected?.updatedAt]
  )

  const loadData = useCallback(async () => {
    const [workflowPage, specs] = await Promise.all([
      fetchAIWorkflows({ page: 1, limit: 50, status: 0 }),
      fetchAIWorkflowNodeSpecs(),
    ])
    setWorkflows(workflowPage?.results ?? [])
    setNodeSpecs(specs ?? [])
  }, [])

  useEffect(() => {
    void loadData().catch((error) => {
      toast.error(error instanceof Error ? error.message : "Failed to load workflows")
    })
  }, [loadData])

  const selectWorkflow = (workflow: AIWorkflow) => {
    setSelected(workflow)
    setName(workflow.name)
    setDescription(workflow.description)
    setOwnerId(String(workflow.ownerId || 1))
    setDefinition(workflow.draftDefinition ?? emptyDefinition)
    setValidation(null)
  }

  const createNew = () => {
    setSelected(null)
    setName("Customer support flow")
    setDescription("")
    setOwnerId("1")
    setDefinition(emptyDefinition)
    setValidation(null)
  }

  const saveDraft = async () => {
    setLoading(true)
    try {
      const payload = {
        name,
        description,
        ownerType: "ai_agent",
        ownerId: Number(ownerId) || 0,
        definition,
      }
      if (selected) {
        await updateAIWorkflow({ id: selected.id, ...payload })
        toast.success("Draft saved")
      } else {
        const created = await createAIWorkflow(payload)
        setSelected(created)
        toast.success("Workflow created")
      }
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to save workflow")
    } finally {
      setLoading(false)
    }
  }

  const runValidation = async () => {
    setLoading(true)
    try {
      const result = await validateAIWorkflow(definition)
      setValidation(result)
      toast[result.valid ? "success" : "error"](
        result.valid ? "Workflow is valid" : "Workflow has validation errors"
      )
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to validate workflow")
    } finally {
      setLoading(false)
    }
  }

  const publish = async () => {
    if (!selected) {
      toast.error("Save the workflow before publishing.")
      return
    }
    setLoading(true)
    try {
      const version = await publishAIWorkflow(selected.id, definition)
      toast.success(`Published version ${version.version}`)
      await loadData()
    } catch (error) {
      toast.error(error instanceof Error ? error.message : "Failed to publish workflow")
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex h-[calc(100vh-var(--header-height))] min-h-0 flex-col overflow-hidden">
      <div className="flex shrink-0 items-center justify-between border-b px-5 py-3">
        <div className="min-w-0">
          <h1 className="truncate text-base font-semibold">AI Workflows</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Edit and publish customer-service conversation flows.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" onClick={createNew}>
            New
          </Button>
          <Button variant="outline" disabled={loading} onClick={runValidation}>
            <CheckCircle2Icon className="size-4" />
            Validate
          </Button>
          <Button variant="outline" disabled={loading} onClick={saveDraft}>
            <SaveIcon className="size-4" />
            Save draft
          </Button>
          <Button disabled={loading || !selected} onClick={publish}>
            <SendIcon className="size-4" />
            Publish
          </Button>
        </div>
      </div>
      <div className="grid min-h-0 flex-1 grid-cols-[300px_minmax(0,1fr)]">
        <aside className="min-h-0 overflow-y-auto border-r bg-muted/20">
          <div className="space-y-4 border-b p-4">
            <div className="space-y-2">
              <Label htmlFor="workflow-name">Name</Label>
              <Input
                id="workflow-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="workflow-owner">AI Agent ID</Label>
              <Input
                id="workflow-owner"
                type="number"
                min={1}
                value={ownerId}
                onChange={(event) => setOwnerId(event.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="workflow-description">Description</Label>
              <Textarea
                id="workflow-description"
                rows={3}
                value={description}
                onChange={(event) => setDescription(event.target.value)}
              />
            </div>
          </div>
          <div className="p-3">
            <div className="mb-2 text-sm font-medium">Workflows</div>
            <div className="space-y-2">
              {workflows.map((workflow) => (
                <button
                  key={workflow.id}
                  type="button"
                  onClick={() => selectWorkflow(workflow)}
                  className={`w-full rounded-md border px-3 py-2 text-left text-sm hover:bg-muted ${
                    selected?.id === workflow.id ? "border-primary bg-primary/5" : "bg-background"
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="truncate font-medium">{workflow.name}</span>
                    {workflow.publishedVersionId ? (
                      <Badge variant="secondary">Published</Badge>
                    ) : null}
                  </div>
                  <div className="mt-1 truncate text-xs text-muted-foreground">
                    Agent #{workflow.ownerId}
                  </div>
                </button>
              ))}
              {workflows.length === 0 ? (
                <div className="rounded-md border border-dashed p-4 text-sm text-muted-foreground">
                  No workflows yet.
                </div>
              ) : null}
            </div>
          </div>
        </aside>
        <main className="flex min-h-0 flex-col overflow-hidden">
          <div className="flex shrink-0 items-center gap-2 border-b px-4 py-2 text-sm">
            <GitBranchIcon className="size-4 text-muted-foreground" />
            <span className="font-medium">{selected ? selected.name : "Unsaved workflow"}</span>
            {validation ? (
              <Badge variant={validation.valid ? "default" : "destructive"}>
                {validation.valid ? "Backend valid" : `${validation.errors.length} backend errors`}
              </Badge>
            ) : null}
            {validation && !validation.valid ? (
              <span className="truncate text-xs text-destructive">
                {validation.errors.map((item) => item.message).join("; ")}
              </span>
            ) : null}
          </div>
          <div className="min-h-0 flex-1">
            <WorkflowEditor
              key={editorKey}
              definition={definition}
              nodeSpecs={nodeSpecs}
              onDefinitionChange={setDefinition}
            />
          </div>
        </main>
      </div>
    </div>
  )
}
