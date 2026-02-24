package io.albertocavalcante.jenkins.steprpc

import com.google.protobuf.Timestamp
import io.albertocavalcante.jenkins.steprpc.v1.Error
import io.albertocavalcante.jenkins.steprpc.v1.RunStatusResponse
import jenkins.model.Jenkins
import org.kohsuke.stapler.HttpResponse

class StepRpcV1RunsApi(private val runStore: InMemoryRunStore) {
    fun doIndex(): HttpResponse {
        Jenkins.get().checkPermission(Jenkins.READ)
        return errorResponse(
            statusCode = 400,
            code = "bad_request",
            message = "run id is required",
        )
    }

    fun getDynamic(runId: String): HttpResponse {
        Jenkins.get().checkPermission(Jenkins.READ)
        val record = runStore.get(runId)
            ?: return errorResponse(
                statusCode = 404,
                code = "run_not_found",
                message = "no run found for id '$runId'",
            )

        val response = RunStatusResponse.newBuilder()
            .setRequestId(record.requestId)
            .setRunId(record.runId)
            .setOperation(record.operation)
            .setState(record.state)
            .setCreatedAt(
                Timestamp.newBuilder()
                    .setSeconds(record.createdAt.epochSecond)
                    .setNanos(record.createdAt.nano)
                    .build(),
            )

        if (record.errorCode != null) {
            response.setError(
                Error.newBuilder()
                    .setCode(record.errorCode)
                    .setMessage(record.errorMessage ?: "operation execution failed")
                    .build(),
            )
        }

        return jsonResponse(response.build())
    }
}
