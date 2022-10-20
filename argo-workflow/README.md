# Argo Workflow User define Sink
This Sink will create Argo workflow with every event. Event content will be pass as `workflow parameter` to workflow.
User can define their workflow define as `workflowtemplate` and configure the `workflowtemplate` name in `env`. This sink
supports caching and dedupe triggered event. you can define the  dedup `keys`, cache `size`, cache `life`.

## Environment Variables

	ARGO_WORKFLOW_TEMPLATE   : Argo WorkflowTemplate name 
	WORKFLOW_NAMESPACE       : Namespace 
	PARAMETER_NAME           : Parameter name for event content
	WORKFLOW_SERVICE_ACCOUNT : Workflow Service Account
	MSG_DEDUP_KEYS           : Message dedup keys
	DEDUP_CACHE_LIMIT        : Cache Size
	DEDUP_CACHE_TTL_DURATION : TTL for each element in cache
	READ_INTERVAL_DURATION   : Workflow Triggering interval
	WORKFLOW_NAME_PREFIX     : Workflow Name prefix

