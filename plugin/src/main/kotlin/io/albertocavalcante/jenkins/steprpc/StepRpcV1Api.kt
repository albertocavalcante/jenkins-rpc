package io.albertocavalcante.jenkins.steprpc

import com.google.protobuf.InvalidProtocolBufferException
import io.albertocavalcante.jenkins.steprpc.v1.CatalogOperation
import io.albertocavalcante.jenkins.steprpc.v1.CatalogResponse
import io.albertocavalcante.jenkins.steprpc.v1.Error
import io.albertocavalcante.jenkins.steprpc.v1.HealthResponse
import io.albertocavalcante.jenkins.steprpc.v1.InvokeRequest
import io.albertocavalcante.jenkins.steprpc.v1.InvokeResponse
import io.albertocavalcante.jenkins.steprpc.v1.OperationExecutionMode
import jenkins.model.Jenkins
import org.kohsuke.stapler.HttpResponse
import org.kohsuke.stapler.StaplerRequest2
import org.kohsuke.stapler.interceptor.RequirePOST

class StepRpcV1Api(
    private val runStore: InMemoryRunStore,
    private val operationRegistry: OperationRegistry,
    private val executor: InRunOperationExecutor,
    private val cpsBridgeQueue: CpsBridgeQueue,
) {
    fun doIndex(): HttpResponse {
        Jenkins.get().checkPermission(Jenkins.READ)
        return jsonResponse(
            HealthResponse.newBuilder()
                .setApiVersion("v1")
                .setService("jenkins-step-rpc-plugin")
                .setStatus("ok")
                .build(),
        )
    }

    fun doCatalog(): HttpResponse {
        Jenkins.get().checkPermission(Jenkins.READ)
        val discovered = executor.discoverOperations()
        val operations = operationRegistry.catalog(discovered).map {
            CatalogOperation.newBuilder()
                .setName(it.name)
                .setDescription(it.description)
                .setExecutionMode(
                    when (it.executionMode) {
                        ExecutionMode.DIRECT -> OperationExecutionMode.OPERATION_EXECUTION_MODE_DIRECT
                        ExecutionMode.CPS_BRIDGE_REQUIRED -> OperationExecutionMode.OPERATION_EXECUTION_MODE_CPS_BRIDGE_REQUIRED
                    },
                )
                .build()
        }
        return jsonResponse(
            CatalogResponse.newBuilder()
                .addAllOperations(operations)
                .build(),
        )
    }

    @RequirePOST
    fun doInvoke(req: StaplerRequest2): HttpResponse {
        Jenkins.get().checkPermission(StepRpcPermissions.INVOKE)

        val body = req.reader.readText()
        if (body.isBlank()) {
            return errorResponse(
                statusCode = 400,
                code = "bad_request",
                message = "request body is required",
            )
        }

        val payload = InvokeRequest.newBuilder()
        try {
            mergeJsonIntoBuilder(body, payload)
        } catch (_: InvalidProtocolBufferException) {
            return errorResponse(
                statusCode = 400,
                code = "bad_json",
                message = "request body must be valid JSON",
            )
        }

        val requestId = payload.requestId
        val operation = payload.operation

        if (requestId.isBlank() || operation.isBlank()) {
            return errorResponse(
                statusCode = 400,
                code = "bad_request",
                message = "requestId and operation are required",
            )
        }

        if (!operationRegistry.isAllowed(operation)) {
            return errorResponse(
                statusCode = 400,
                code = "operation_not_allowed",
                message = "operation '$operation' is not in allowlist",
            )
        }

        val args = structToAnyMap(payload.args)
        val redactedArgs = AuditLogger.redactSensitiveArgs(args)
        AuditLogger.log(
            "invoke.start",
            mapOf("requestId" to requestId, "operation" to operation, "args" to redactedArgs.toString()),
        )

        val execution = executor.execute(
            requestId = requestId,
            operation = operation,
            args = args,
        )

        val record = runStore.create(
            requestId = requestId,
            runId = execution.runId,
            operation = operation,
            state = execution.state,
            errorCode = execution.errorCode,
            errorMessage = execution.errorMessage,
        )

        AuditLogger.log(
            "invoke.complete",
            mapOf("requestId" to requestId, "operation" to operation, "runId" to record.runId, "state" to record.state),
        )

        val responseBuilder = InvokeResponse.newBuilder()
            .setRequestId(record.requestId)
            .setRunId(record.runId)
            .setState(record.state)
        if (record.errorCode != null) {
            responseBuilder.setError(
                Error.newBuilder()
                    .setCode(record.errorCode)
                    .setMessage(record.errorMessage ?: "operation execution failed")
                    .build(),
            )
        }

        return jsonResponse(
            responseBuilder.build(),
        )
    }

    fun getRuns(): StepRpcV1RunsApi = StepRpcV1RunsApi(runStore)

    fun getBridge(): StepRpcV1BridgeApi = StepRpcV1BridgeApi(runStore, cpsBridgeQueue)
}
