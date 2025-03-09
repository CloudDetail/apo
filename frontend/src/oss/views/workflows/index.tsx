import WorkflowsIframe from './workflowsIframe'

const WorkflowsPage = () => {
  return (
    <div className="text-xs" style={{ height: 'calc(100vh - 120px)' }}>
      <WorkflowsIframe src="/dify/apps" />
    </div>
  )
}

export default WorkflowsPage
