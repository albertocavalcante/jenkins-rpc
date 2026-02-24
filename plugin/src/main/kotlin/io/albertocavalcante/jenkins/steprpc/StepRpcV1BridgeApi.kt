package io.albertocavalcante.jenkins.steprpc

import com.google.protobuf.InvalidProtocolBufferException
import com.google.protobuf.Struct
import io.albertocavalcante.jenkins.steprpc.v1.BridgeCompleteRequest
import io.albertocavalcante.jenkins.steprpc.v1.BridgeCompleteResponse
import io.albertocavalcante.jenkins.steprpc.v1.BridgePendingResponse
import jenkins.model.Jenkins
import org.kohsuke.stapler.HttpResponse
import org.kohsuke.stapler.StaplerRequest2
import org.kohsuke.stapler.interceptor.RequirePOST

class StepRpcV1BridgeApi(
    private val runStore: InMemoryRunStore,
    private val cpsBridgeQueue: CpsBridgeQueue,
) {
    fun doPending(req: StaplerRequest2): HttpResponse {
        Jenkins.get().checkPermission(StepRpcPermissions.INVOKE)
        val targetRun = req.getParameter("runExternalizableId")
        if (targetRun.isNullOrBlank()) {
            return errorResponse(
                statusCode = 400,
                code = "bad_request",
                message = "runExternalizableId query parameter is required",
            )
        }

        val pending = cpsBridgeQueue.nextForRun(targetRun)
            ?: return errorResponse(
                statusCode = 404,
                code = "no_pending_request",
                message = "no pending bridge request for run '$targetRun'",
            )

        AuditLogger.log(
            "bridge.pending",
            mapOf("targetRun" to targetRun, "runId" to pending.runId, "operation" to pending.operation),
        )

        val argsBuilder = Struct.newBuilder()
        mergeJsonIntoBuilder(net.sf.json.JSONObject.fromObject(pending.args).toString(), argsBuilder)

        return jsonResponse(
            BridgePendingResponse.newBuilder()
                .setRequestId(pending.requestId)
                .setRunId(pending.runId)
                .setOperation(pending.operation)
                .setArgs(argsBuilder.build())
                .setTargetRunExternalizableId(pending.targetRunExternalizableId)
                .build(),
        )
    }

    @RequirePOST
    fun doComplete(req: StaplerRequest2): HttpResponse {
        Jenkins.get().checkPermission(StepRpcPermissions.INVOKE)
        val body = req.reader.readText()
        if (body.isBlank()) {
            return errorResponse(
                statusCode = 400,
                code = "bad_request",
                message = "request body is required",
            )
        }

        val payload = BridgeCompleteRequest.newBuilder()
        try {
            mergeJsonIntoBuilder(body, payload)
        } catch (_: InvalidProtocolBufferException) {
            return errorResponse(
                statusCode = 400,
                code = "bad_json",
                message = "request body must be valid JSON",
            )
        }

        val runId = payload.runId
        val state = payload.state
        if (runId.isBlank() || state.isBlank()) {
            return errorResponse(
                statusCode = 400,
                code = "bad_request",
                message = "runId and state are required",
            )
        }

        if (state != "succeeded" && state != "failed" && state != "cancelled") {
            return errorResponse(
                statusCode = 400,
                code = "bad_request",
                message = "state must be one of succeeded, failed, cancelled",
            )
        }

        val completed = cpsBridgeQueue.complete(runId)
            ?: return errorResponse(
                statusCode = 404,
                code = "run_not_found",
                message = "no pending bridge request found for run '$runId'",
            )

        val errorCode = if (payload.hasError() && payload.error.code.isNotBlank()) payload.error.code else null
        val errorMessage = if (payload.hasError() && payload.error.message.isNotBlank()) payload.error.message else null
        runStore.update(
            runId = completed.runId,
            state = state,
            errorCode = errorCode,
            errorMessage = errorMessage,
        )

        AuditLogger.log(
            "bridge.complete",
            mapOf("runId" to completed.runId, "state" to state),
        )

        return jsonResponse(
            BridgeCompleteResponse.newBuilder()
                .setRequestId(completed.requestId)
                .setRunId(completed.runId)
                .setState(state)
                .build(),
        )
    }
}
