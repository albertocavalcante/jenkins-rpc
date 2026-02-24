# Bridge Worker Example (Pipeline Side)

This snippet shows one way to execute queued CPS-bound operations from inside a running Pipeline build.

```groovy
def rpcBase = "${env.JENKINS_URL}step-rpc/v1"
def runExternalizableId = currentBuild.rawBuild.externalizableId

def drainOnce = {
  def pendingResp = httpRequest(
    url: "${rpcBase}/bridge/pending?runExternalizableId=${java.net.URLEncoder.encode(runExternalizableId, 'UTF-8')}",
    validResponseCodes: '200,404'
  )
  if (pendingResp.status == 404) {
    return false
  }

  def pending = readJSON text: pendingResp.content
  def state = "succeeded"
  def err = null

  try {
    this."${pending.operation}"(pending.args ?: [:])
  } catch (Exception ex) {
    state = "failed"
    err = [code: "operation_failed", message: ex.message ?: "bridge execution failed"]
  }

  def completePayload = [runId: pending.runId, state: state]
  if (err != null) {
    completePayload.error = err
  }
  httpRequest(
    httpMode: 'POST',
    contentType: 'APPLICATION_JSON',
    url: "${rpcBase}/bridge/complete",
    requestBody: groovy.json.JsonOutput.toJson(completePayload),
    validResponseCodes: '200'
  )
  return true
}

while (drainOnce()) {
  // keep draining until queue is empty for this run
}
```

Notes:

1. This is a minimal example for trusted Pipeline usage.
2. Production setup should add auth headers and stricter error handling.
3. For long-running jobs, invoke the drain loop at well-defined checkpoints.
