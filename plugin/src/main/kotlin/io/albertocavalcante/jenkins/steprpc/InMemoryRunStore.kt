package io.albertocavalcante.jenkins.steprpc

import java.time.Instant
import java.util.concurrent.ConcurrentHashMap

data class RunRecord(
    val requestId: String,
    val runId: String,
    val operation: String,
    val state: String,
    val createdAt: Instant,
    val errorCode: String? = null,
    val errorMessage: String? = null,
)

class InMemoryRunStore {
    private val byRunID = ConcurrentHashMap<String, RunRecord>()

    fun put(record: RunRecord) {
        byRunID[record.runId] = record
    }

    fun get(runId: String): RunRecord? = byRunID[runId]

    fun create(
        requestId: String,
        runId: String,
        operation: String,
        state: String,
        errorCode: String? = null,
        errorMessage: String? = null,
    ): RunRecord {
        val record = RunRecord(
            requestId = requestId,
            runId = runId,
            operation = operation,
            state = state,
            createdAt = Instant.now(),
            errorCode = errorCode,
            errorMessage = errorMessage,
        )
        put(record)
        return record
    }

    fun update(
        runId: String,
        state: String,
        errorCode: String? = null,
        errorMessage: String? = null,
    ): RunRecord? {
        val current = byRunID[runId] ?: return null
        val updated = current.copy(
            state = state,
            errorCode = errorCode,
            errorMessage = errorMessage,
        )
        put(updated)
        return updated
    }
}
